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
	Path       string
	URL        string
	RemoteSha1 string
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

// FetchAndRead updates the cache with data from cache.URL, if set, and reads the contents of the cache into v.
func (cache JSONCache[T]) FetchAndRead(v *T) error {
	if cache.URL == "" {
		return fmt.Errorf("no URL to fetch from")
	}
	if err := DownloadFile(DownloadEntry{
		URL:  cache.URL,
		Path: cache.Path,
		Sha1: cache.RemoteSha1,
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

type Cache struct {
	Path string
	URL  string
}

// Read reads the contents of the cache into v.
func (cache Cache) Read(v *[]byte) error {
	data, err := os.ReadFile(cache.Path)
	if err != nil {
		return err
	}
	*v = data
	return nil
}

// FetchAndRead updates the cache with data from cache.URL, if set, and reads the contents of the cache into v.
func (cache Cache) FetchAndRead(v *[]byte) error {
	if cache.URL == "" {
		return fmt.Errorf("no URL to fetch from")
	}
	if err := DownloadFile(DownloadEntry{
		URL:  cache.URL,
		Path: cache.Path,
	}); err != nil {
		return err
	}

	return cache.Read(v)
}
