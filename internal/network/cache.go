package network

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

var ErrNotCached = errors.New("data not cached and request failed")

// A Cache stores and retrieves remote data to unmarshal either into JSON or a custom unmarshaler.
type Cache[T any] struct {
	Path        string
	URL         string
	RemoteSha1  string
	AlwaysFetch bool
	Unmarshaler func(data []byte, v any) error // Custom unmarshal function. Defaults to JSON.
}

// Get checks the cache and checks if it is valid. If it is, its contents are returned. If not, they are fetched and then returned.
func (cache Cache[T]) Get(v *T) error {
	download := true
	if _, err := os.Stat(cache.Path); err == nil {
		if cache.RemoteSha1 != "" {
			sum, err := cache.Sha1()
			if err != nil {
				return err
			}
			if cache.RemoteSha1 == "" || cache.RemoteSha1 == sum {
				download = false
			}
		}
	}

	if download || cache.AlwaysFetch {
		if cache.URL == "" {
			return fmt.Errorf("no URL to fetch from")
		}

		err := DownloadFile(DownloadEntry{
			URL:  cache.URL,
			Path: cache.Path,
			Sha1: cache.RemoteSha1,
		})
		if err != nil && download {
			return fmt.Errorf("%w: %w", ErrNotCached, err)
		}
	}

	data, err := os.ReadFile(cache.Path)
	if err != nil {
		return err
	}

	if cache.Unmarshaler != nil {
		return cache.Unmarshaler(data, v)
	} else {
		return json.Unmarshal(data, v)
	}
}

// Sha1 returns the SHA1 checksum of the cache
func (cache Cache[T]) Sha1() (string, error) {
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
