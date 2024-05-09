package util

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
)

func CheckResponse(resp *http.Response, err error) error {
	if err != nil {
		return fmt.Errorf("Request could not be created: %v", err)
	}
	if resp.StatusCode > 299 || resp.StatusCode < 200 {
		return fmt.Errorf("Request to %v failed with status %v", resp.Request.URL, resp.Status)
	}
	return nil
}

func GetJSON(url string, v any) error {
	resp, err := http.Get(url)
	if err := CheckResponse(resp, err); err != nil {
		return err
	}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, v)
	return nil
}

func DownloadFile(url string, dest string) error {
	if _, err := os.Stat(dest); err == nil {
		return nil
	}
	fmt.Println("Downloading file: " + url)
	err := os.MkdirAll(path.Dir(dest), os.ModePerm)
	if err != nil {
		return fmt.Errorf("Failed to create directory structure for file: %v", err)
	}
	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("Failed to create file %v: %v", dest, err)
	}
	defer out.Close()
	resp, err := http.Get(url)
	if err := CheckResponse(resp, err); err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(out, resp.Body)
	return nil
}
