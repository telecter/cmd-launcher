package launcher

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	util "github.com/telecter/cmd-launcher/internal"
	"github.com/telecter/cmd-launcher/internal/api"
)

func getPathFromMaven(mavenPath string) string {
	identifier := strings.Split(mavenPath, ":")
	groupID := strings.Replace(identifier[0], ".", "/", -1)
	basename := fmt.Sprintf("%s-%s.jar", identifier[1], identifier[2])
	return fmt.Sprintf("%s/%s/%s/%s", groupID, identifier[1], identifier[2], basename)
}

func getLibraries(libraries []api.Library, rootDir string) ([]string, error) {
	var artifacts []api.Artifact
	for _, library := range libraries {
		install := true
		for _, rule := range library.Rules {
			if rule.Os.Name == "" {
				continue
			}
			os := rule.Os.Name
			os = strings.ReplaceAll(os, "osx", "darwin")
			if os != runtime.GOOS && rule.Action == "allow" {
				install = false
			}
			if os == runtime.GOOS && rule.Action == "disallow" {
				install = false
			}
		}
		if !install {
			continue
		}
		classifiers := library.Downloads.Classifiers
		osToClassifier := map[string]api.Artifact{
			"darwin":  classifiers.NativesMacOS,
			"linux":   classifiers.NativesLinux,
			"windows": classifiers.NativesWindows,
		}

		if classifier, ok := osToClassifier[runtime.GOOS]; ok && classifier.URL != "" {
			artifacts = append(artifacts, classifier)
		}

		if library.URL != "" {
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
		path := filepath.Join(rootDir, "libraries", artifact.Path)
		err := util.DownloadFile(artifact.URL, path)
		if err != nil {
			return paths, fmt.Errorf("error while downloading libraries: %s", err)
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func getAssets(meta api.VersionMeta, rootDir string) error {
	index := api.AssetIndex{}
	indexPath := filepath.Join(rootDir, "assets", "indexes", meta.AssetIndex.ID+".json")

	if data, err := os.ReadFile(indexPath); err == nil {
		json.Unmarshal(data, &index)
	} else {
		if err := util.GetJSON(meta.AssetIndex.URL, &index); err != nil {
			return fmt.Errorf("error while getting asset index: %s", err)
		}
	}
	for _, asset := range index.Objects {
		prefix := asset.Hash[:2]
		util.DownloadFile(fmt.Sprintf("https://resources.download.minecraft.net/%s/%s", prefix, asset.Hash), filepath.Join(rootDir, "assets", "objects", prefix, asset.Hash))
	}

	util.DownloadFile(meta.AssetIndex.URL, indexPath)
	return nil
}
