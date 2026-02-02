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

func (s *FilePersistenceSystem) Flush(ctx context.Context, w *GameWorld) error {
	if s.baseDir == "" {
		return fmt.Errorf("file persistence: baseDir is empty")
	}
	return w.Entities.ForEach(ctx, func(ent ginka_ecs_go.DataEntity) error {
		dirtyTypes := ent.DirtyTypes()
		if len(dirtyTypes) == 0 {
			return nil
		}
		persisted := make([]ginka_ecs_go.ComponentType, 0, len(dirtyTypes))
		for _, t := range dirtyTypes {
			component, ok := ent.Get(t)
			if !ok {
				continue
			}
			dataComponent, ok := component.(ginka_ecs_go.DataComponent)
			if !ok {
				return fmt.Errorf("file persistence: component %d is not a DataComponent", t)
			}
			marshaler, ok := component.(interface{ Marshal() ([]byte, error) })
			if !ok {
				return fmt.Errorf("file persistence: component %d does not implement Marshal", t)
			}
			payload, err := marshaler.Marshal()
			if err != nil {
				return err
			}
			key := sanitizeKey(dataComponent.StorageKey())
			path := filepath.Join(s.baseDir, ent.Id(), key+".json")
			if err := writeFileAtomic(path, payload, 0o644); err != nil {
				return err
			}
			persisted = append(persisted, t)
		}
		if len(persisted) > 0 {
			ent.ClearDirty(persisted...)
		}
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
