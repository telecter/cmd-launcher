package launcher

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

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
	err := network.DownloadFile(artifact.URL, library.RuntimePath())
	if err != nil {
		return fmt.Errorf("download artifact '%s': %w", artifact.Path, err)
	}
	return nil
}

func (library RuntimeLibrary) RuntimePath() string {
	return filepath.Join(internal.LibrariesDir, library.Artifact().Path)
}

func filterLibraries(libraries []meta.Library) (installed []RuntimeLibrary, required []RuntimeLibrary) {
	for _, library := range libraries {
		library := RuntimeLibrary{library}
		if library.ShouldBeInstalled() {
			if runtime.GOOS == "linux" && runtime.GOARCH == "arm64" && strings.HasPrefix(library.Name, "org.lwjgl") {
				path := strings.ReplaceAll(library.Downloads.Artifact.Path, "linux", "linux-arm64")
				library = RuntimeLibrary{meta.GetMavenLibrary(library.Name, path)}
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

func getClientLibrary(versionMeta meta.VersionMeta) meta.Library {
	return meta.Library{
		Name: "com.mojang:minecraft:" + versionMeta.ID,
		Downloads: struct {
			Artifact meta.Artifact "json:\"artifact\""
		}{
			Artifact: meta.Artifact{
				Path: fmt.Sprintf("com/mojang/minecraft/%s/%s.jar", versionMeta.ID, versionMeta.ID),
				Sha1: versionMeta.Downloads.Client.Sha1,
				Size: versionMeta.Downloads.Client.Size,
				URL:  versionMeta.Downloads.Client.URL,
			},
		}}
}

func fixLibraries(libraries []meta.Library, loader Loader) []meta.Library {
	if loader == LoaderFabric {
		for i, library := range libraries {
			if strings.Contains(library.Name, "org.ow2.asm:asm:") {
				libraries = slices.Delete(libraries, i, i+1)
				break
			}
		}
	}
	return libraries
}
