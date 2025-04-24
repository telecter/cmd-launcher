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
	"github.com/telecter/cmd-launcher/internal/meta"
	"github.com/telecter/cmd-launcher/internal/network"
)

// for Fabric libraries
func getPathFromMaven(mavenPath string) string {
	identifier := strings.Split(mavenPath, ":")
	groupID := strings.Replace(identifier[0], ".", "/", -1)
	basename := fmt.Sprintf("%s-%s.jar", identifier[1], identifier[2])
	return fmt.Sprintf("%s/%s/%s/%s", groupID, identifier[1], identifier[2], basename)
}

func installLibraries(libraries []meta.Library) ([]string, error) {
	var artifacts []meta.Artifact
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
						artifacts = append(artifacts, meta.Artifact{
							Path: artifactPath,
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
			artifacts = append(artifacts, meta.Artifact{
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

func downloadAssets(versionMeta meta.VersionMeta) error {
	var index meta.AssetIndex
	indexPath := filepath.Join(env.RootDir, "assets", "indexes", versionMeta.AssetIndex.ID+".json")

	downloadAssetIndex := true
	if data, err := os.ReadFile(indexPath); err == nil {
		downloadAssetIndex = false
		if err := json.Unmarshal(data, &index); err != nil {
			log.Println("Current asset index invalid. Downloading a new one.")
			downloadAssetIndex = true
		}
	}
	if downloadAssetIndex {
		if err := network.FetchJSONData(versionMeta.AssetIndex.URL, &index); err != nil {
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
		network.DownloadFile(versionMeta.AssetIndex.URL, indexPath)
	}
	return nil
}
