package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

func (s *FilePersistenceSystem) Flush(ctx context.Context, w ginka_ecs_go.World) error {
	if s.baseDir == "" {
		return fmt.Errorf("file persistence: baseDir is empty")
	}
	return w.Entities().ForEach(ctx, func(ent ginka_ecs_go.DataEntity) error {
		dirty := ent.DirtyDataComponents()
		if len(dirty) == 0 {
			return nil
		}
		for _, c := range dirty {
			payload, err := c.Marshal()
			if err != nil {
				return err
			}
			key := sanitizeKey(c.StorageKey())
			path := filepath.Join(s.baseDir, ent.Id(), key+".json")
			if err := writeFileAtomic(path, payload, 0o644); err != nil {
				return err
			}
		}
		ent.ClearDirty()
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
