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

	"github.com/telecter/cmd-launcher/internal/meta"
	"github.com/telecter/cmd-launcher/internal/network"
	env "github.com/telecter/cmd-launcher/pkg"
)

type runtimeLibrary struct {
	meta.Library
	artifact    meta.Artifact
	runtimePath string
}

func newRuntimeLibrary(library meta.Library) runtimeLibrary {
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

	return runtimeLibrary{
		Library:     library,
		artifact:    artifact,
		runtimePath: filepath.Join(env.LibrariesDir, artifact.Path),
	}
}

func (library runtimeLibrary) shouldBeInstalled() bool {
	if len(library.Rules) > 0 {
		rule := library.Rules[0]
		os := strings.ReplaceAll(rule.Os.Name, "osx", "darwin")
		return os == runtime.GOOS && rule.Action == "allow"
	}
	return true
}

func (library runtimeLibrary) isInstalled() bool {
	data, err := os.ReadFile(filepath.Join(env.LibrariesDir, library.artifact.Path))
	if err != nil {
		return false
	}
	// if no checksum is present, still count the library as installed as long as the file exists
	if library.artifact.Sha1 == "" {
		return true
	}
	sum := sha1.Sum(data)
	return library.artifact.Sha1 == hex.EncodeToString(sum[:])
}

func (library runtimeLibrary) downloadEntry() network.DownloadEntry {
	return network.DownloadEntry{
		URL:      library.artifact.URL,
		Filename: library.runtimePath,
	}
}

func filterLibraries(libraries []meta.Library) (installed []runtimeLibrary, required []runtimeLibrary) {
	for _, library := range libraries {
		library := newRuntimeLibrary(library)
		if library.shouldBeInstalled() {
			if runtime.GOOS == "linux" && runtime.GOARCH == "arm64" && strings.HasPrefix(library.Name, "org.lwjgl") {
				path := strings.ReplaceAll(library.Downloads.Artifact.Path, "linux", "linux-arm64")
				library = newRuntimeLibrary(meta.GetMavenLibrary(library.Name, path))
			}
			if library.isInstalled() {
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
