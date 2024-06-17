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
	var paths []string
	for _, library := range libraries {
		if len(library.Rules) > 0 {
			os := library.Rules[0].Os.Name
			os = strings.ReplaceAll(os, "osx", "darwin")
			if os != "" && os != runtime.GOOS {
				continue
			}
		}
		var url string
		var file string
		if library.URL != "" {
			url = library.URL + getPathFromMaven(library.Name)
			file = filepath.Join(rootDir, "libraries", getPathFromMaven(library.Name))
		} else {
			url = library.Downloads.Artifact.URL
			file = filepath.Join(rootDir, "libraries", library.Downloads.Artifact.Path)
		}
		err := util.DownloadFile(url, file)
		if err != nil {
			return paths, fmt.Errorf("error while downloading %s: %s", library.Name, err)
		}
		paths = append(paths, file)
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
