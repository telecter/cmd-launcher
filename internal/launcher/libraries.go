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
	Artifact    meta.Artifact
	RuntimePath string
}

func NewRuntimeLibrary(library meta.Library) RuntimeLibrary {
	artifact := library.Downloads.Artifact
	if library.URL != "" {
		identifier := strings.Split(library.Name, ":")
		group := strings.ReplaceAll(identifier[0], ".", "/")
		path := strings.Join([]string{group, identifier[1], identifier[2], fmt.Sprintf("%s-%s.jar", identifier[1], identifier[2])}, "/")
		artifact = meta.Artifact{
			Path: path,
			URL:  library.URL + "/" + path,
			Sha1: library.Sha1,
			Size: library.Size,
		}
	}

	return RuntimeLibrary{
		Library:     library,
		Artifact:    artifact,
		RuntimePath: filepath.Join(internal.LibrariesDir, artifact.Path),
	}
}

func (library RuntimeLibrary) ShouldBeInstalled() bool {
	if len(library.Rules) > 0 {
		rule := library.Rules[0]
		os := strings.ReplaceAll(rule.Os.Name, "osx", "darwin")
		return os == runtime.GOOS && rule.Action == "allow"
	}
	return true
}

func (library RuntimeLibrary) IsInstalled() bool {
	data, err := os.ReadFile(filepath.Join(internal.LibrariesDir, library.Artifact.Path))
	if err != nil {
		return false
	}
	// if no checksum is present, still count the library as installed as long as the file exists
	if library.Artifact.Sha1 == "" {
		return true
	}
	sum := sha1.Sum(data)
	return library.Artifact.Sha1 == hex.EncodeToString(sum[:])
}

func (library RuntimeLibrary) DownloadEntry() network.DownloadEntry {
	return network.DownloadEntry{
		URL:      library.Artifact.URL,
		Filename: library.RuntimePath,
	}
}

func filterLibraries(libraries []meta.Library) (installed []RuntimeLibrary, required []RuntimeLibrary) {
	for _, library := range libraries {
		library := NewRuntimeLibrary(library)
		if library.ShouldBeInstalled() {
			if runtime.GOOS == "linux" && runtime.GOARCH == "arm64" && strings.HasPrefix(library.Name, "org.lwjgl") {
				path := strings.ReplaceAll(library.Downloads.Artifact.Path, "linux", "linux-arm64")
				library = NewRuntimeLibrary(meta.GetMavenLibrary(library.Name, path))
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
