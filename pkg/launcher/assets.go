package launcher

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/telecter/cmd-launcher/internal/meta"
	"github.com/telecter/cmd-launcher/internal/network"
	env "github.com/telecter/cmd-launcher/pkg"
)

// An asset is a runtime representation of a game asset.
type asset struct {
	meta.AssetObject
	URL         string
	RuntimePath string
}

// newAsset creates an asset from a meta.AssetObject.
func newAsset(object meta.AssetObject) asset {
	return asset{
		URL:         fmt.Sprintf(meta.MINECRAFT_RESOURCES_URL, object.Hash[:2], object.Hash),
		RuntimePath: filepath.Join(env.AssetsDir, "objects", object.Hash[:2], object.Hash),
		AssetObject: object,
	}
}

// isDownloaded reports whether asset exists and has a valid checksum.
func (asset asset) isDownloaded() bool {
	data, err := os.ReadFile(asset.RuntimePath)
	if err != nil {
		return false
	}
	sum := sha1.Sum(data)
	return asset.Hash == hex.EncodeToString(sum[:])
}

// downloadEntry returns a DownloadEntry to fetch asset.
func (asset asset) downloadEntry() network.DownloadEntry {
	return network.DownloadEntry{
		URL:      asset.URL,
		Filename: asset.RuntimePath,
	}
}

// filterAssets takes index, transforms its objects to assets and returns the ones that are not downloaded.
func filterAssets(index meta.AssetIndex) (required []asset) {
	for _, object := range index.Objects {
		asset := newAsset(object)
		if !asset.isDownloaded() {
			required = append(required, asset)
		}
	}
	return required
}

// downloadAssetIndex retrieves the asset index for the specified version.
func downloadAssetIndex(versionMeta meta.VersionMeta) (meta.AssetIndex, error) {
	cache := network.JSONCache[meta.AssetIndex]{
		Path: filepath.Join(env.AssetsDir, "indexes", versionMeta.AssetIndex.ID+".json"),
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
