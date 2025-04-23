package launcher

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/telecter/cmd-launcher/internal/env"
	"github.com/telecter/cmd-launcher/internal/network"
	"github.com/telecter/cmd-launcher/internal/network/api"
)

// for Fabric libraries
func getPathFromMaven(mavenPath string) string {
	identifier := strings.Split(mavenPath, ":")
	groupID := strings.Replace(identifier[0], ".", "/", -1)
	basename := fmt.Sprintf("%s-%s.jar", identifier[1], identifier[2])
	return fmt.Sprintf("%s/%s/%s/%s", groupID, identifier[1], identifier[2], basename)
}

func installLibraries(libraries []api.Library) ([]string, error) {
	var artifacts []api.Artifact
	for _, library := range libraries {
		if len(library.Rules) > 0 { // library has rules
			shouldInstall := false

			for _, rule := range library.Rules {
				os := rule.Os.Name
				os = strings.ReplaceAll(os, "osx", "darwin")

				if os == runtime.GOOS && rule.Action == "allow" {
					shouldInstall = true
				}
				// is system running arm64 linux
				if os == runtime.GOOS && rule.Action == "allow" && runtime.GOOS == "linux" && runtime.GOARCH == "arm64" {
					if strings.HasPrefix(library.Name, "org.lwjgl") {
						artifactPath := strings.ReplaceAll(library.Downloads.Artifact.Path, ".jar", "-arm64.jar")
						artifacts = append(artifacts, api.Artifact{
							Path: library.Downloads.Artifact.Path,
							URL:  "https://repo1.maven.org/maven2/" + artifactPath,
						})
						shouldInstall = false
					}

				}
			}

			if !shouldInstall {
				continue
			}
		}

		if library.URL != "" {
			// duplicate fabric asm library
			if library.Name == "org.ow2.asm:asm:9.8" {
				continue
			}
			artifacts = append(artifacts, api.Artifact{
				Path: getPathFromMaven(library.Name),
				URL:  library.URL + getPathFromMaven(library.Name),
			})
		} else {
			artifacts = append(artifacts, library.Downloads.Artifact)
		}
	}

	var paths []string
	for _, artifact := range artifacts {
		path := filepath.Join(env.RootDir, "libraries", artifact.Path)
		err := network.DownloadFile(artifact.URL, path)
		if err != nil {
			return paths, fmt.Errorf("error while downloading libraries: %w", err)
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func downloadAssets(meta api.VersionMeta) error {
	var index api.AssetIndex
	indexPath := filepath.Join(env.RootDir, "assets", "indexes", meta.AssetIndex.ID+".json")

	downloadAssetIndex := true
	if data, err := os.ReadFile(indexPath); err == nil {
		downloadAssetIndex = false
		if err := json.Unmarshal(data, &index); err != nil {
			log.Println("Current asset index invalid. Downloading a new one.")
			downloadAssetIndex = true
		}
	}
	if downloadAssetIndex {
		if err := network.FetchJSONData(meta.AssetIndex.URL, &index); err != nil {
			return fmt.Errorf("error while downloading asset index: %w", err)
		}
	}

	for _, asset := range index.Objects {
		loc, _ := url.JoinPath("https://resources.download.minecraft.net", asset.Hash[:2], asset.Hash)
		if err := network.DownloadFile(loc, filepath.Join(env.RootDir, "assets", "objects", asset.Hash[:2], asset.Hash)); err != nil {
			log.Println("Warning! Asset download failed.")
		}
	}

	if downloadAssetIndex {
		os.Remove(indexPath)
		network.DownloadFile(meta.AssetIndex.URL, indexPath)
	}
	return nil
}

func downloadClient(meta api.VersionMeta, name string) error {
	if err := network.DownloadFile(meta.Downloads.Client.URL, filepath.Join(env.RootDir, "versions", name, name+".jar")); err != nil {
		return fmt.Errorf("error downloading client: %s", err)
	}
	return nil
}
