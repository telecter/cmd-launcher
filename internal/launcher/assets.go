package launcher

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/telecter/cmd-launcher/internal"
	"github.com/telecter/cmd-launcher/internal/meta"
	"github.com/telecter/cmd-launcher/internal/network"
)

func filterAssets(index meta.AssetIndex) meta.AssetIndex {
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

func downloadAsset(asset meta.AssetObject) error {
	url := fmt.Sprintf(meta.MINECRAFT_RESOURCES_URL, asset.Hash[:2], asset.Hash)
	if err := network.DownloadFile(url, filepath.Join(internal.AssetsDir, "objects", asset.Hash[:2], asset.Hash)); err != nil {
		return fmt.Errorf("download asset: %w", err)
	}
	return nil
}

func downloadAssetIndex(versionMeta meta.VersionMeta) (meta.AssetIndex, error) {
	cache := network.JSONCache[meta.AssetIndex]{
		Path: filepath.Join(internal.AssetsDir, "indexes", versionMeta.AssetIndex.ID+".json"),
		URL:  versionMeta.AssetIndex.URL,
	}
	download := true

	var assetIndex meta.AssetIndex
	if err := cache.Read(&assetIndex); err == nil {
		sum, _ := cache.Sha1()
		if sum == versionMeta.AssetIndex.Sha1 {
			download = false
		}
	}
	if download {
		if err := cache.UpdateAndRead(&assetIndex); err != nil {
			return meta.AssetIndex{}, fmt.Errorf("fetch asset index: %w", err)
		}
	}
	return assetIndex, nil
}
