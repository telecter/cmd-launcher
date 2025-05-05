package launcher

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/schollz/progressbar/v3"
	"github.com/telecter/cmd-launcher/internal"
	"github.com/telecter/cmd-launcher/internal/meta"
	"github.com/telecter/cmd-launcher/internal/network"
)

type RuntimeLibrary struct {
	meta.Library
}

func (library RuntimeLibrary) IsFabric() bool {
	return library.URL != ""
}
func (library RuntimeLibrary) Artifact() meta.Artifact {
	if !library.IsFabric() {
		return library.Downloads.Artifact
	}
	identifier := strings.Split(library.Name, ":")
	group := strings.ReplaceAll(identifier[0], ".", "/")
	path := strings.Join([]string{group, identifier[1], identifier[2], fmt.Sprintf("%s-%s.jar", identifier[1], identifier[2])}, "/")

	return meta.Artifact{
		Path: path,
		URL:  strings.Join([]string{library.URL, path}, "/"),
		Sha1: library.Sha1,
		Size: library.Size,
	}
}
func (library RuntimeLibrary) ShouldBeInstalled() bool {
	install := true
	if len(library.Rules) > 0 {
		install = false
		rule := library.Rules[0]
		os := rule.Os.Name
		os = strings.ReplaceAll(os, "osx", "darwin")
		if os == runtime.GOOS && rule.Action == "allow" {
			install = true
		}
	}
	return install
}
func (library RuntimeLibrary) IsInstalled() bool {
	artifact := library.Artifact()
	data, err := os.ReadFile(filepath.Join(internal.LibrariesDir, artifact.Path))
	if err != nil {
		return false
	}
	// if no checksum is present, still count the library as installed as long as the file exists
	if artifact.Sha1 == "" {
		return true
	}
	sum := sha1.Sum(data)
	return artifact.Sha1 == hex.EncodeToString(sum[:])
}
func (library RuntimeLibrary) Install() error {
	artifact := library.Artifact()
	err := network.DownloadFile(artifact.URL, filepath.Join(internal.LibrariesDir, artifact.Path))
	if err != nil {
		return fmt.Errorf("download artifact '%s': %w", artifact.Path, err)
	}
	return nil
}
func (library RuntimeLibrary) RuntimePath() string {
	return filepath.Join(internal.LibrariesDir, library.Artifact().Path)
}

func fetchLibraryFromMaven(name string, path string) (RuntimeLibrary, error) {
	url := fmt.Sprintf("https://repo1.maven.org/maven2/%s", path)
	checksumCachePath := filepath.Join(internal.CachesDir, filepath.Base(path)+".sha1")
	var checksum []byte
	checksum, err := os.ReadFile(checksumCachePath)
	if err != nil {
		resp, err := http.Get(fmt.Sprintf("%s.sha1", url))
		if err != nil {
			return RuntimeLibrary{}, fmt.Errorf("get Maven library checksum: %w", err)
		}
		defer resp.Body.Close()
		checksum, _ = io.ReadAll(resp.Body)
		if err := os.WriteFile(checksumCachePath, checksum, 0644); err != nil {
			return RuntimeLibrary{}, fmt.Errorf("cache Maven library checksum: %w", err)
		}
	}

	return RuntimeLibrary{meta.Library{
		Name: name,
		Downloads: struct {
			Artifact meta.Artifact "json:\"artifact\""
		}{
			Artifact: meta.Artifact{
				Path: path,
				URL:  url,
				Sha1: string(checksum),
			},
		},
	}}, nil
}

func filterLibraries(libraries []meta.Library) (installed []RuntimeLibrary, required []RuntimeLibrary) {
	for _, library := range libraries {
		library := RuntimeLibrary{library}
		if library.ShouldBeInstalled() {
			if runtime.GOOS == "linux" && runtime.GOARCH == "arm64" && strings.HasPrefix(library.Name, "org.lwjgl") {
				path := strings.ReplaceAll(library.Downloads.Artifact.Path, "linux", "linux-arm64")
				library, _ = fetchLibraryFromMaven(library.Name, path)
			}
			if library.IsInstalled() {
				installed = append(installed, library)
			} else {
				required = append(required, library)
			}
		}
	}
	return installed, required
}

func getRuntimeLibraryPaths(libraries []RuntimeLibrary) (paths []string) {
	for _, library := range libraries {
		paths = append(paths, library.RuntimePath())
	}
	return paths
}

func installLibraries(libraries []RuntimeLibrary) error {
	if len(libraries) < 1 {
		return nil
	}
	bar := progressbar.Default(int64(len(libraries)), "Installing libraries")
	for _, library := range libraries {
		if err := library.Install(); err != nil {
			return fmt.Errorf("download library '%s': %w", library.Name, err)
		}
		bar.Add(1)
	}

	return nil
}
