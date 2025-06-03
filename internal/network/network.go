package network

import (
	"encoding/json"
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
	Filename string
}

// FetchJSON fetches url and unmarshals its JSON contents into v
func FetchJSON(url string, v any) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return json.Unmarshal(data, v)
}

// DownloadFile downloads the specified DownloadEntry and saves it.
//
// All parent directories are created in order to create the file.
func DownloadFile(entry DownloadEntry) error {
	resp, err := http.Get(entry.URL)
	if err != nil {
		return err
	}
	if err := CheckResponse(resp); err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := os.MkdirAll(filepath.Dir(entry.Filename), 0755); err != nil {
		return fmt.Errorf("create directory for file '%s': %w", entry.Filename, err)
	}
	out, err := os.Create(entry.Filename)
	if err != nil {
		return fmt.Errorf("create file '%s': %w", entry.Filename, err)
	}
	defer out.Close()
	io.Copy(out, resp.Body)
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

// CheckResponse ensures the status code of an http.Response is successful.
func CheckResponse(resp *http.Response) error {
	if resp.StatusCode > 299 || resp.StatusCode < 200 {
		return fmt.Errorf("%s %s (%s)", resp.Request.Method, resp.Request.URL, resp.Status)
	}
	return nil
}
