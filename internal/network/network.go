package network

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

func FetchJSONData(url string, v any) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return json.Unmarshal(data, v)
}

func DownloadFile(url string, dest string) error {
	if _, err := os.Stat(dest); err == nil {
		return nil
	}
	err := os.MkdirAll(path.Dir(dest), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory for file: %s", err)
	}
	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %s", dest, err)
	}
	defer out.Close()
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode > 299 || resp.StatusCode < 200 {
		return err
	}
	defer resp.Body.Close()
	log.Printf("Downloading...%s\n", url)
	io.Copy(out, resp.Body)
	return nil
}

func CheckResponse(resp *http.Response, err error) error {
	if err != nil {
		return err
	}
	if resp.StatusCode > 299 || resp.StatusCode < 200 {
		return fmt.Errorf("%s %s (%s)", resp.Request.Method, resp.Request.URL, resp.Status)
	}
	return nil
}
