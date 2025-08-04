package network_test

import (
	"net/http"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/telecter/cmd-launcher/internal/meta"
	"github.com/telecter/cmd-launcher/internal/network"
)

func TestDownloadFile(t *testing.T) {
	tests := []struct {
		name      string
		entry     network.DownloadEntry
		wantError bool
	}{
		{
			name: "Normal",
			entry: network.DownloadEntry{
				URL:  "https://resources.download.minecraft.net/5f/5ff04807c356f1beed0b86ccf659b44b9983e3fa",
				Path: filepath.Join(t.TempDir(), "normal.png"),
			},
			wantError: false,
		},
		{
			name: "SHA1",
			entry: network.DownloadEntry{
				URL:  "https://resources.download.minecraft.net/5f/5ff04807c356f1beed0b86ccf659b44b9983e3fa",
				Sha1: "5ff04807c356f1beed0b86ccf659b44b9983e3fa",
				Path: filepath.Join(t.TempDir(), "with_sha1.png"),
			},
			wantError: false,
		},
		{
			name: "Invalid URL",
			entry: network.DownloadEntry{
				URL:  "https://resources.upload.minecraft.net/5f/5ff04807c356f1beed0b86ccf659b44b9983e3fa",
				Path: filepath.Join(t.TempDir(), "with_invalid_url.png"),
			},
			wantError: true,
		},
		{
			name: "Invalid SHA1",
			entry: network.DownloadEntry{
				URL:  "https://resources.download.minecraft.net/5f/5ff04807c356f1beed0b86ccf659b44b9983e3fa",
				Path: filepath.Join(t.TempDir(), "with_invalid_sha1.png"),
				Sha1: "not valid",
			},
			wantError: true,
		},
		{
			name: "Bad Path",
			entry: network.DownloadEntry{
				URL:  "https://resources.download.minecraft.net/5f/5ff04807c356f1beed0b86ccf659b44b9983e3fa",
				Path: "/nopermission",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := network.DownloadFile(tt.entry)
			isError := err != nil
			if tt.wantError != isError {
				t.Errorf("got error: %t; wanted error to be: %t", isError, tt.wantError)
				if isError {
					t.Log(err)
				}
			}
		})
	}
}

func TestStartDownloadEntries(t *testing.T) {
	results := network.StartDownloadEntries([]network.DownloadEntry{
		{
			URL:  "https://resources.download.minecraft.net/5f/5ff04807c356f1beed0b86ccf659b44b9983e3fa",
			Path: filepath.Join(t.TempDir(), "normal.png"),
		},
	})
	for err := range results {
		if err != nil {
			t.Errorf("wanted no error; got: %s", err)
		}
	}
}

func TestCheckResponse_OK(t *testing.T) {
	err := network.CheckResponse(&http.Response{StatusCode: 200})
	if err != nil {
		t.Errorf("wanted no error; got: %s", err)
	}
}

func TestCheckResponse_Error(t *testing.T) {
	u, _ := url.Parse("https://telecter.xyz")

	err := network.CheckResponse(&http.Response{
		StatusCode: 404,
		Request: &http.Request{
			URL:    u,
			Method: http.MethodGet,
		},
	})
	if err == nil {
		t.Errorf("wanted error; got no error")
	}
}

func genCache(tempDir string) network.Cache[meta.VersionMeta] {
	return network.Cache[meta.VersionMeta]{
		Path:       filepath.Join(tempDir, "meta.json"),
		URL:        "https://piston-meta.mojang.com/v1/packages/24b08e167c6611f7ad895ae1e8b5258f819184aa/1.21.8.json",
		RemoteSha1: "24b08e167c6611f7ad895ae1e8b5258f819184aa",
	}
}

func TestCache_Read(t *testing.T) {
	cache := genCache(t.TempDir())

	var data meta.VersionMeta
	if err := cache.Read(&data); err != nil {
		t.Errorf("wanted no error; got: %s", err)
	}
}

func TestCache_Sha1(t *testing.T) {
	cache := genCache(t.TempDir())

	var data meta.VersionMeta
	if err := cache.Read(&data); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	hash, err := cache.Sha1()
	if err != nil {
		t.Errorf("wanted no error; got: %s", err)
	}
	if len(hash) != 40 {
		t.Error("wanted checksum; got empty string")
	}
}
