package meta

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/telecter/cmd-launcher/internal"
)

// Paths are relative to env.CachesDir

type JSONCache struct {
	Path string
}

func (cache JSONCache) GetAbsolutePath() string {
	return filepath.Join(internal.CachesDir, cache.Path)
}

func (cache JSONCache) Read(v any) error {
	data, err := os.ReadFile(cache.GetAbsolutePath())
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

	if err := os.MkdirAll(filepath.Dir(cache.GetAbsolutePath()), 0755); err != nil {
		return fmt.Errorf("failed to create directory for cache: %w", err)
	}

	if err := os.WriteFile(cache.GetAbsolutePath(), data, 0644); err != nil {
		return fmt.Errorf("failed to write cache: %w", err)
	}
	return nil
}

func (cache JSONCache) Sha1Sum() (string, error) {
	data, err := os.ReadFile(cache.GetAbsolutePath())
	if err != nil {
		return "", fmt.Errorf("failed to read cache: %w", err)
	}
	sum := sha1.Sum(data)
	return string(sum[:]), nil
}
