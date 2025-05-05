package network

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type JSONCache struct {
	Path string
}

func (cache JSONCache) Read(v any) error {
	data, err := os.ReadFile(cache.Path)
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
		return fmt.Errorf("marshal data: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(cache.Path), 0755); err != nil {
		return fmt.Errorf("create directory for cache: %w", err)
	}

	if err := os.WriteFile(cache.Path, data, 0644); err != nil {
		return fmt.Errorf("write cache: %w", err)
	}
	return nil
}

func (cache JSONCache) Sha1Sum() (string, error) {
	data, err := os.ReadFile(cache.Path)
	if err != nil {
		return "", fmt.Errorf("read cache: %w", err)
	}
	sum := sha1.Sum(data)
	return string(sum[:]), nil
}
