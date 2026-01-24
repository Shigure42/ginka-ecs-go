package ginka_ecs_go

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type worldRuntime struct {
	shards    []worldShard
	shardMask uint64

	stopping atomic.Bool
	refs     atomic.Int64
	refsMu   sync.Mutex
	refsCond *sync.Cond

	cmdAll []CommandSystem
	cmdBy  map[CommandType][]CommandSystem

	shardedTickSystems []ShardedTickSystem
	tickCancel         context.CancelFunc

	tickWG sync.WaitGroup
}

type worldShard struct {
	ch chan shardRequest
	wg sync.WaitGroup
}

type shardRequestKind int

const (
	shardRequestCommand shardRequestKind = iota + 1
	shardRequestTick
)

type shardRequest struct {
	kind shardRequestKind

	ctx context.Context

	cmd Command
	dt  time.Duration

	shardIdx   int
	shardCount int

	resp chan error
}

func (w *CoreWorld) finalizeLocked() (*worldRuntime, error) {
	if w.shardCount <= 0 {
		w.shardCount = defaultShardCount
	}
	count := nextPow2(uint64(w.shardCount))
	if count == 0 {
		return nil, fmt.Errorf("invalid shard count")
	}
	mask := count - 1

	rt := &worldRuntime{
		shards:             make([]worldShard, int(count)),
		shardMask:          mask,
		cmdAll:             append([]CommandSystem(nil), w.commandAll...),
		cmdBy:              make(map[CommandType][]CommandSystem, len(w.commandBy)),
		shardedTickSystems: append([]ShardedTickSystem(nil), w.shardedTickSystems...),
	}
	rt.refsCond = sync.NewCond(&rt.refsMu)
	for typ, list := range w.commandBy {
		rt.cmdBy[typ] = append([]CommandSystem(nil), list...)
	}
	return rt, nil
}

func (rt *worldRuntime) tryAcquire() bool {
	if rt.stopping.Load() {
		return false
	}
	rt.refs.Add(1)
	if rt.stopping.Load() {
		rt.release()
		return false
	}
	return true
}

func (rt *worldRuntime) release() {
	if rt.refs.Add(-1) != 0 {
		return
	}
	rt.refsMu.Lock()
	rt.refsCond.Broadcast()
	rt.refsMu.Unlock()
}

func (w *CoreWorld) startLocked(rt *worldRuntime) {
	for i := range rt.shards {
		rt.shards[i].ch = make(chan shardRequest, 256)
		rt.shards[i].wg.Add(1)
		go func(shard *worldShard, shardIdx int, shardCount int) {
			defer shard.wg.Done()
			for req := range shard.ch {
				var err error
				switch req.kind {
				case shardRequestCommand:
					err = dispatchCommand(req.ctx, w, rt, req.cmd)
				case shardRequestTick:
					err = dispatchTickShard(req.ctx, w, rt, req.dt, shardIdx, shardCount)
				default:
					err = fmt.Errorf("unknown shard request kind %d", req.kind)
				}
				req.resp <- err
			}
		}(&rt.shards[i], i, len(rt.shards))
	}

	ctx, cancel := context.WithCancel(context.Background())
	rt.tickCancel = cancel

	interval := w.tickInterval
	if interval <= 0 {
		return
	}
	handler := w.tickErrHandler

	rt.tickWG.Add(1)
	go func() {
		defer rt.tickWG.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		last := time.Now()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				dt := now.Sub(last)
				last = now
				if err := w.TickOnce(ctx, dt); err != nil {
					if handler != nil {
						handler(err)
					}
				}
			}
		}
	}()
}

func (rt *worldRuntime) tickOnce(ctx context.Context, dt time.Duration, _ World) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	shardCount := len(rt.shards)
	if shardCount == 0 {
		return nil
	}

	resps := make([]chan error, shardCount)
	for i := 0; i < shardCount; i++ {
		resp := make(chan error, 1)
		resps[i] = resp
		req := shardRequest{
			kind:       shardRequestTick,
			ctx:        ctx,
			dt:         dt,
			shardIdx:   i,
			shardCount: shardCount,
			resp:       resp,
		}
		select {
		case rt.shards[i].ch <- req:
			// sent to shard for serial execution
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	for i := 0; i < shardCount; i++ {
		select {
		case err := <-resps[i]:
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (rt *worldRuntime) stop() {
	rt.stopping.Store(true)

	if rt.tickCancel != nil {
		rt.tickCancel()
		rt.tickWG.Wait()
	}

	rt.refsMu.Lock()
	for rt.refs.Load() != 0 {
		rt.refsCond.Wait()
	}
	rt.refsMu.Unlock()

	for i := range rt.shards {
		if rt.shards[i].ch != nil {
			close(rt.shards[i].ch)
		}
	}
	for i := range rt.shards {
		rt.shards[i].wg.Wait()
	}
}

func dispatchCommand(ctx context.Context, w World, rt *worldRuntime, cmd Command) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	all := rt.cmdAll
	by := rt.cmdBy[cmd.Type()]
	if len(all) == 0 && len(by) == 0 {
		return fmt.Errorf("submit command %d: %w", cmd.Type(), ErrUnhandledCommand)
	}

	for _, sys := range all {
		if err := sys.Handle(ctx, w, cmd); err != nil {
			return err
		}
	}
	for _, sys := range by {
		if err := sys.Handle(ctx, w, cmd); err != nil {
			return err
		}
	}
	return nil
}

func dispatchTickShard(ctx context.Context, w World, rt *worldRuntime, dt time.Duration, shardIdx int, shardCount int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, sys := range rt.shardedTickSystems {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := sys.TickShard(ctx, w, dt, shardIdx, shardCount); err != nil {
			return err
		}
	}
	return nil
}
