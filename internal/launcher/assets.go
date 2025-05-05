package launcher

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/schollz/progressbar/v3"
	"github.com/telecter/cmd-launcher/internal"
	"github.com/telecter/cmd-launcher/internal/meta"
	"github.com/telecter/cmd-launcher/internal/network"
)

func getRequiredAssets(index meta.AssetIndex) meta.AssetIndex {
	var assets meta.AssetIndex
	assets.Objects = make(map[string]meta.AssetObject)
	for i, object := range index.Objects {
		data, err := os.ReadFile(filepath.Join(internal.AssetsDir, "objects", object.Hash[:2], object.Hash))
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
	for name, asset := range index.Objects {
		url := fmt.Sprintf("https://resources.download.minecraft.net/%s/%s", asset.Hash[:2], asset.Hash)
		if err := network.DownloadFile(url, filepath.Join(internal.AssetsDir, "objects", asset.Hash[:2], asset.Hash)); err != nil {
			return fmt.Errorf("download asset '%s': %w", name, err)
		}
		bar.Add(1)
	}
	return nil
}

func downloadAssetIndex(versionMeta meta.VersionMeta) (meta.AssetIndex, error) {
	cache := network.JSONCache{Path: filepath.Join(internal.AssetsDir, "indexes", versionMeta.AssetIndex.ID+".json")}

	var assetIndex meta.AssetIndex
	if err := cache.Read(&assetIndex); err != nil {
		if err := network.FetchJSON(versionMeta.AssetIndex.URL, &assetIndex); err != nil {
			return meta.AssetIndex{}, fmt.Errorf("fetch asset index: %w", err)
		}
		cache.Write(assetIndex)
	}
	return assetIndex, nil
}
