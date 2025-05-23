package network

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type JSONCache[T any] struct {
	Path string
	URL  string
}

func (cache JSONCache[T]) Read(v *T) error {
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
func (cache JSONCache[T]) UpdateAndRead(v *T) error {
	if err := DownloadFile(cache.URL, cache.Path); err != nil {
		return fmt.Errorf("update cache: %w", err)
	}
	if err := cache.Read(v); err != nil {
		return fmt.Errorf("read cache: %w", err)
	}
	return nil
}
func (cache JSONCache[T]) Sha1() (string, error) {
	f, err := os.Open(cache.Path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
