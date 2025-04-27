package launcher

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/schollz/progressbar/v3"
	"github.com/telecter/cmd-launcher/internal/env"
	"github.com/telecter/cmd-launcher/internal/meta"
	"github.com/telecter/cmd-launcher/internal/network"
)

func getRequiredAssets(index meta.AssetIndex) meta.AssetIndex {
	var assets meta.AssetIndex
	assets.Objects = make(map[string]meta.AssetObject)
	for i, object := range index.Objects {
		data, err := os.ReadFile(filepath.Join(env.AssetsDir, "objects", object.Hash[:2], object.Hash))
		if err != nil {
			assets.Objects[i] = object
		}
		sum := sha1.Sum(data)
		if object.Hash != hex.EncodeToString(sum[:]) {
			assets.Objects[i] = object
		}
	}
	return assets
}

func downloadAssets(index meta.AssetIndex) error {
	if len(index.Objects) < 1 {
		return nil
	}
	bar := progressbar.Default(int64(len(index.Objects)), "Downloading assets")
	for _, asset := range index.Objects {
		url := fmt.Sprintf("https://resources.download.minecraft.net/%s/%s", asset.Hash[:2], asset.Hash)
		if err := network.DownloadFile(url, filepath.Join(env.AssetsDir, "objects", asset.Hash[:2], asset.Hash)); err != nil {
			log.Println("Warning! Asset download failed.")
		}
		bar.Add(1)
	}
	return nil
}

func downloadAssetIndex(versionMeta meta.VersionMeta) (meta.AssetIndex, error) {
	path := filepath.Join(env.AssetsDir, "indexes", versionMeta.AssetIndex.ID+".json")

	var assetIndex meta.AssetIndex
	data, err := os.ReadFile(path)
	if err != nil {
		if err := network.FetchJSON(versionMeta.AssetIndex.URL, &assetIndex); err != nil {
			return meta.AssetIndex{}, fmt.Errorf("failed to fetch asset index: %w", err)
		}
		data, _ := json.Marshal(assetIndex)
		os.MkdirAll(filepath.Dir(path), 0755)
		if err := os.WriteFile(path, data, 0644); err != nil {
			panic(err)
		}
	} else {
		if err := json.Unmarshal(data, &assetIndex); err != nil {
			return meta.AssetIndex{}, fmt.Errorf("failed to read cached asset index: %w", err)
		}
	}
	return assetIndex, nil
}
