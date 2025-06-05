package network_test

import (
	"path/filepath"
	"testing"

	"github.com/telecter/cmd-launcher/internal/network"
)

func TestDownload(t *testing.T) {
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
			name: "With SHA1",
			entry: network.DownloadEntry{
				URL:  "https://resources.download.minecraft.net/5f/5ff04807c356f1beed0b86ccf659b44b9983e3fa",
				Sha1: "5ff04807c356f1beed0b86ccf659b44b9983e3fa",
				Path: filepath.Join(t.TempDir(), "with_sha1.png"),
			},
			wantError: false,
		},
		{
			name: "With Invalid URL",
			entry: network.DownloadEntry{
				URL:  "https://resources.upload.minecraft.net/5f/5ff04807c356f1beed0b86ccf659b44b9983e3fa",
				Path: filepath.Join(t.TempDir(), "with_invalid_url.png"),
			},
			wantError: true,
		},
		{
			name: "With Invalid SHA1",
			entry: network.DownloadEntry{
				URL:  "https://resources.download.minecraft.net/5f/5ff04807c356f1beed0b86ccf659b44b9983e3fa",
				Path: filepath.Join(t.TempDir(), "with_invalid_sha1.png"),
				Sha1: "not valid",
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
