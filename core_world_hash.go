package ginka_ecs_go

// hashEntityId mixes an entity id for shard routing.
//
// This is intentionally cheap; it does not need to be cryptographically strong.
func hashEntityId(id uint64) uint64 {
	// A splitmix64-style mix.
	x := id + 0x9e3779b97f4a7c15
	x = (x ^ (x >> 30)) * 0xbf58476d1ce4e5b9
	x = (x ^ (x >> 27)) * 0x94d049bb133111eb
	x = x ^ (x >> 31)
	return x
}

func nextPow2(n uint64) uint64 {
	if n == 0 {
		return 1
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	return n + 1
}

// ShardIndex returns the shard index for an entity id.
//
// The returned value matches CoreWorld's internal sharding algorithm.
//
// Note: shardCount is expected to be the actual runtime shard count passed to
// ShardedTickSystem.TickShard (which is a power of two). If a non-power-of-two
// shardCount is provided, it is rounded up to the next power of two.
func ShardIndex(entityId uint64, shardCount int) int {
	if shardCount <= 0 {
		return 0
	}
	n := uint64(shardCount)
	if n&(n-1) != 0 {
		n = nextPow2(n)
	}
	return int(hashEntityId(entityId) & (n - 1))
}
