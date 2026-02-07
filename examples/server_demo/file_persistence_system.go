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

type persistedComponent struct {
	typ     ginka_ecs_go.ComponentType
	path    string
	payload []byte
	version uint64
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
	if err := ctx.Err(); err != nil {
		return err
	}
	createdDirs := make(map[string]struct{})
	return w.Entities.ForEach(ctx, func(ent ginka_ecs_go.DataEntity) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		persisted := make([]persistedComponent, 0)
		missing := make([]ginka_ecs_go.ComponentType, 0)
		if err := ent.Tx(func(tx ginka_ecs_go.DataEntity) error {
			dirtyTypes := tx.DirtyTypes()
			if len(dirtyTypes) == 0 {
				return nil
			}
			persisted = make([]persistedComponent, 0, len(dirtyTypes))
			missing = make([]ginka_ecs_go.ComponentType, 0, len(dirtyTypes))
			for _, t := range dirtyTypes {
				if err := ctx.Err(); err != nil {
					return err
				}
				component, ok := tx.Get(t)
				if !ok {
					missing = append(missing, t)
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
				persisted = append(persisted, persistedComponent{
					typ:     t,
					path:    filepath.Join(s.baseDir, tx.Id(), key+".json"),
					payload: payload,
					version: dataComponent.Version(),
				})
			}
			return nil
		}); err != nil {
			return err
		}

		for _, item := range persisted {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := writeFileAtomic(item.path, item.payload, 0o644, createdDirs); err != nil {
				return err
			}
		}

		if len(persisted) == 0 && len(missing) == 0 {
			return nil
		}

		if err := ent.Tx(func(tx ginka_ecs_go.DataEntity) error {
			toClear := make(map[ginka_ecs_go.ComponentType]struct{}, len(persisted)+len(missing))
			for _, t := range missing {
				if _, ok := tx.Get(t); !ok {
					toClear[t] = struct{}{}
				}
			}
			for _, item := range persisted {
				component, ok := tx.Get(item.typ)
				if !ok {
					continue
				}
				dataComponent, ok := component.(ginka_ecs_go.DataComponent)
				if !ok {
					continue
				}
				if dataComponent.Version() == item.version {
					toClear[item.typ] = struct{}{}
				}
			}
			if len(toClear) == 0 {
				return nil
			}
			types := make([]ginka_ecs_go.ComponentType, 0, len(toClear))
			for t := range toClear {
				types = append(types, t)
			}
			tx.ClearDirty(types...)
			return nil
		}); err != nil {
			return err
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

func writeFileAtomic(path string, data []byte, perm os.FileMode, createdDirs map[string]struct{}) error {
	dir := filepath.Dir(path)
	if _, ok := createdDirs[dir]; !ok {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", dir, err)
		}
		createdDirs[dir] = struct{}{}
	}

	tmpFile, err := os.CreateTemp(dir, filepath.Base(path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file in %s: %w", dir, err)
	}
	tmpPath := tmpFile.Name()
	defer func() {
		_ = os.Remove(tmpPath)
	}()

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("write %s: %w", tmpPath, err)
	}
	if err := tmpFile.Chmod(perm); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("chmod %s: %w", tmpPath, err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close %s: %w", tmpPath, err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename %s -> %s: %w", tmpPath, path, err)
	}
	return nil
}
