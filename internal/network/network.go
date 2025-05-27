package network

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
)

func FetchJSON(url string, v any) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return json.Unmarshal(data, v)
}

func DownloadFile(url string, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if err := CheckResponse(resp); err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := os.MkdirAll(path.Dir(dest), 0755); err != nil {
		return fmt.Errorf("create directory for file '%s': %w", dest, err)
	}
	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("create file '%s': %w", dest, err)
	}
	defer out.Close()
	io.Copy(out, resp.Body)
	return nil
}

func CheckResponse(resp *http.Response) error {
	if resp.StatusCode > 299 || resp.StatusCode < 200 {
		return fmt.Errorf("%s %s (%s)", resp.Request.Method, resp.Request.URL, resp.Status)
	}
	return nil
}
