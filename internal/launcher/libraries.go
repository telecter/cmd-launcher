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
	"github.com/telecter/cmd-launcher/internal/env"
	"github.com/telecter/cmd-launcher/internal/meta"
	"github.com/telecter/cmd-launcher/internal/network"
)

type RuntimeLibrary struct {
	meta.Library
}

func (library RuntimeLibrary) IsFabric() bool {
	return library.URL != ""
}
func (library RuntimeLibrary) GetArtifact() meta.Artifact {
	if !library.IsFabric() {
		return library.Downloads.Artifact
	}
	return meta.Artifact{
		Path: mavenNameToPath(library.Name),
		URL:  fmt.Sprintf("%s/%s", library.URL, mavenNameToPath(library.Name)),
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
	artifact := library.GetArtifact()
	data, err := os.ReadFile(filepath.Join(env.LibrariesDir, artifact.Path))
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
	artifact := library.GetArtifact()
	err := network.DownloadFile(artifact.URL, filepath.Join(env.LibrariesDir, artifact.Path))
	if err != nil {
		return fmt.Errorf("error while downloading artifact '%s': %w", artifact.Path, err)
	}
	return nil
}
func (library RuntimeLibrary) GetRuntimePath() (string, error) {
	if !library.IsInstalled() {
		return "", fmt.Errorf("library is not installed")
	}
	return filepath.Join(env.LibrariesDir, library.GetArtifact().Path), nil
}

// for Fabric libraries
func mavenNameToPath(mavenPath string) string {
	identifier := strings.Split(mavenPath, ":")
	groupID := strings.Replace(identifier[0], ".", "/", -1)
	basename := fmt.Sprintf("%s-%s.jar", identifier[1], identifier[2])
	return fmt.Sprintf("%s/%s/%s/%s", groupID, identifier[1], identifier[2], basename)
}

func fetchLibraryFromMaven(name string, path string) (RuntimeLibrary, error) {
	url := fmt.Sprintf("https://repo1.maven.org/maven2/%s", path)

	checksumCachePath := filepath.Join(env.CachesDir, filepath.Base(path)+".sha1")
	var checksum []byte
	checksum, err := os.ReadFile(checksumCachePath)
	if err != nil {
		resp, err := http.Get(fmt.Sprintf("%s.sha1", url))
		if err != nil {
			return RuntimeLibrary{}, fmt.Errorf("failed to get Maven library checksum: %w", err)
		}
		defer resp.Body.Close()
		checksum, _ = io.ReadAll(resp.Body)
		if err := os.WriteFile(checksumCachePath, checksum, 0644); err != nil {
			return RuntimeLibrary{}, fmt.Errorf("failed to cache Maven library checksum: %w", err)
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
				//log.Printf("IS INSTALLED: %s\n", library.Name)
				installed = append(installed, library)
			} else {
				//log.Printf("INSTALLING: %s\n", library.Name)
				required = append(required, library)
			}
		}
	}
	return installed, required
}

func getRuntimeLibraryPaths(libraries []RuntimeLibrary) (paths []string) {
	for _, library := range libraries {
		path := filepath.Join(env.LibrariesDir, library.GetArtifact().Path)
		paths = append(paths, path)
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
			return fmt.Errorf("error while downloading library '%s': %w", library.Name, err)
		}
		bar.Add(1)
	}

	return nil
}
