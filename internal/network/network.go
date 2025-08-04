package network

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

const MAX_CONCURRENT_DOWNLOADS = 6

type DownloadEntry struct {
	URL      string
	Path     string
	Sha1     string
	FileMode os.FileMode
}

// DownloadFile downloads the specified DownloadEntry and saves it.
//
// All parent directories are created in order to create the file.
func DownloadFile(entry DownloadEntry) error {
	resp, err := http.Get(entry.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := CheckResponse(resp); err != nil {
		return err
	}
	fmt.Println(filepath.Dir(entry.Path))
	if err := os.MkdirAll(filepath.Dir(entry.Path), 0755); err != nil {
		return fmt.Errorf("create directory for file %q: %w", entry.Path, err)
	}
	out, err := os.Create(entry.Path)
	if err != nil {
		return fmt.Errorf("create file %q: %w", entry.Path, err)
	}
	defer out.Close()

	if entry.FileMode != 0 {
		if err := out.Chmod(entry.FileMode); err != nil {
			return fmt.Errorf("set permissions for file %q: %w", entry.Path, err)
		}
	}

	hash := sha1.New()
	tee := io.TeeReader(resp.Body, hash)

	if _, err := io.Copy(out, tee); err != nil {
		return err
	}

	if entry.Sha1 != "" {
		if hex.EncodeToString(hash.Sum(nil)) != entry.Sha1 {
			return fmt.Errorf("invalid checksum from %q", entry.URL)
		}
	}

	return nil
}

// StartDownloadEntries runs DownloadFile on each specified DownloadEntry and returns a channel with the download results.
func StartDownloadEntries(entries []DownloadEntry) chan error {
	var wg sync.WaitGroup
	results := make(chan error)
	d := make(chan struct{}, MAX_CONCURRENT_DOWNLOADS)
	for _, entry := range entries {
		wg.Add(1)
		go func(entry DownloadEntry) {
			defer wg.Done()

			d <- struct{}{}
			err := DownloadFile(entry)
			<-d
			results <- err
		}(entry)
	}
	go func() {
		wg.Wait()
		close(results)
	}()
	return results
}

type HTTPStatusError struct {
	URL        string
	Method     string
	StatusCode int
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("%s %s (%d)", e.Method, e.URL, e.StatusCode)
}

// CheckResponse ensures the status code of an http.Response is successful, returning an HTTPStatusError if not.
func CheckResponse(resp *http.Response) error {
	if resp.StatusCode > 299 || resp.StatusCode < 200 {
		return &HTTPStatusError{
			URL:        resp.Request.URL.String(),
			Method:     resp.Request.Method,
			StatusCode: resp.StatusCode,
		}
	}
	return nil
}
