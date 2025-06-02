package network

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
)

type JSONCache[T any] struct {
	Path string
	URL  string
}

// Read reads the contents of the cache into v.
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

// UpdateAndRead updates the cache with data from cache.URL and reads the contents of the cache into v.
func (cache JSONCache[T]) UpdateAndRead(v *T) error {
	if err := DownloadFile(DownloadEntry{
		URL:      cache.URL,
		Filename: cache.Path,
	}); err != nil {
		return err
	}
	if err := cache.Read(v); err != nil {
		return err
	}
	return nil
}

// Sha1 returns the SHA1 checksum of the cache
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
