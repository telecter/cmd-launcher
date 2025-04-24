package meta

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/telecter/cmd-launcher/internal/env"
)

// NOTE: All paths relative to env.CachesDir

func isCacheValid(location string) bool {
	path := filepath.Join(env.CachesDir, location)
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}
func readCache(location string, v any) error {
	path := filepath.Join(env.CachesDir, location)
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, v)
	if err != nil {
		return err
	}
	return nil
}
func writeCache(location string, v any) error {
	path := filepath.Join(env.CachesDir, location)
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache: %w", err)
	}
	return nil
}
