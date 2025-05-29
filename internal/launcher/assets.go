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

type Asset struct {
	meta.AssetObject
	URL         string
	RuntimePath string
}

func NewAsset(object meta.AssetObject) Asset {
	return Asset{
		URL:         fmt.Sprintf(meta.MINECRAFT_RESOURCES_URL, object.Hash[:2], object.Hash),
		RuntimePath: filepath.Join(internal.AssetsDir, "objects", object.Hash[:2], object.Hash),
		AssetObject: object,
	}
}

func (asset Asset) IsDownloaded() bool {
	data, err := os.ReadFile(asset.RuntimePath)
	if err != nil {
		return false
	}
	sum := sha1.Sum(data)
	return asset.Hash == hex.EncodeToString(sum[:])
}

func (asset Asset) DownloadEntry() network.DownloadEntry {
	return network.DownloadEntry{
		URL:      asset.URL,
		Filename: asset.RuntimePath,
	}
}

func filterAssets(index meta.AssetIndex) (required []Asset) {
	for _, object := range index.Objects {
		asset := NewAsset(object)
		if !asset.IsDownloaded() {
			required = append(required, asset)
		}
	}
	return required
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
			return meta.AssetIndex{}, err
		}
	}
	return assetIndex, nil
}
