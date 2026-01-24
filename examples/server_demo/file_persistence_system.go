package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	ginka_ecs_go "github.com/Shigure42/ginka-ecs-go"
)

type FilePersistenceSystem struct {
	baseDir string
}

func NewFilePersistenceSystem(baseDir string) *FilePersistenceSystem {
	return &FilePersistenceSystem{baseDir: baseDir}
}

func (s *FilePersistenceSystem) Name() string {
	return "file-persistence"
}

func (s *FilePersistenceSystem) TickShard(ctx context.Context, w ginka_ecs_go.World, _ time.Duration, shardIdx int, shardCount int) error {
	if s.baseDir == "" {
		return fmt.Errorf("file persistence: baseDir is empty")
	}
	return w.Entities().ForEach(ctx, func(ent ginka_ecs_go.DataEntity) error {
		if ginka_ecs_go.ShardIndex(ent.Id(), shardCount) != shardIdx {
			return nil
		}
		dirty := ent.DirtyTypes()
		if len(dirty) == 0 {
			return nil
		}
		for _, t := range dirty {
			c, ok := ent.GetData(t)
			if !ok {
				continue
			}
			payload, err := c.Marshal()
			if err != nil {
				return err
			}
			key := sanitizeKey(c.PersistKey())
			path := filepath.Join(s.baseDir, fmt.Sprintf("%d", ent.Id()), key+".json")
			if err := writeFileAtomic(path, payload, 0o644); err != nil {
				return err
			}
		}
		ent.ClearDirty(dirty...)
		return nil
	})
}

func sanitizeKey(key string) string {
	key = strings.TrimSpace(key)
	key = strings.ReplaceAll(key, "/", "_")
	key = strings.ReplaceAll(key, "\\", "_")
	if key == "" {
		return "component"
	}
	return key
}

func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return fmt.Errorf("write %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename %s -> %s: %w", tmp, path, err)
	}
	return nil
}
