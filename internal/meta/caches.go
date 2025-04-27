package meta

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/telecter/cmd-launcher/internal/env"
)

// Paths are relative to env.CachesDir

type JSONCache struct {
	Path string
}

func getAbsolutePath(path string) string {
	return filepath.Join(env.CachesDir, path)
}

func (cache JSONCache) Read(v any) error {
	data, err := os.ReadFile(getAbsolutePath(cache.Path))
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, v)
	if err != nil {
		return err
	}
	return nil
}
func (cache JSONCache) Write(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(getAbsolutePath(cache.Path)), 0755); err != nil {
		return fmt.Errorf("failed to create directory for cache: %w", err)
	}

	if err := os.WriteFile(getAbsolutePath(cache.Path), data, 0644); err != nil {
		return fmt.Errorf("failed to write cache: %w", err)
	}
	return nil
}
