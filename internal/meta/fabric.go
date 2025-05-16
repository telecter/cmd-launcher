package meta

import (
	"fmt"
	"path/filepath"

	"github.com/telecter/cmd-launcher/internal"
	"github.com/telecter/cmd-launcher/internal/network"
)

type FabricVersionList []struct {
	Loader struct {
		Version string `json:"version"`
	} `json:"loader"`
}
type FabricMeta struct {
	ID           string `json:"id"`
	InheritsFrom string `json:"inheritsFrom"`
	ReleaseTime  string `json:"releaseTime"`
	Time         string `json:"time"`
	Type         string `json:"type"`
	MainClass    string `json:"mainClass"`
	Arguments    struct {
		Game []any    `json:"game"`
		Jvm  []string `json:"jvm"`
	} `json:"arguments"`
	Libraries []Library
}

func GetFabricMeta(gameVersion string) (FabricMeta, error) {
	versionsCache := network.JSONCache{
		Path: filepath.Join(internal.CachesDir, "fabric", fmt.Sprintf("%s-versions.json", gameVersion)),
	}
	profileCache := network.JSONCache{
		Path: filepath.Join(internal.CachesDir, "fabric", fmt.Sprintf("%s-profile.json", gameVersion)),
	}

	var list FabricVersionList
	err := versionsCache.Read(&list)
	if err != nil {
		err = network.FetchJSON(fmt.Sprintf("https://meta.fabricmc.net/v2/versions/loader/%s", gameVersion), &list)
		if err != nil {
			return FabricMeta{}, fmt.Errorf("retrieve Fabric versions: %w", err)
		}
		versionsCache.Write(list)
	}

	var meta FabricMeta
	if err := profileCache.Read(&meta); err != nil {
		err := network.FetchJSON(fmt.Sprintf("https://meta.fabricmc.net/v2/versions/loader/%s/%s/profile/json", gameVersion, list[0].Loader.Version), &meta)
		if err != nil {
			return FabricMeta{}, fmt.Errorf("retrieve metadata for Fabric version %s: %w", list[0].Loader.Version, err)
		}
		profileCache.Write(meta)
	}

	return meta, nil
}

func GetQuiltMeta(gameVersion string) (FabricMeta, error) {
	versionsCache := network.JSONCache{
		Path: filepath.Join(internal.CachesDir, "quilt", fmt.Sprintf("%s-versions.json", gameVersion)),
	}
	profileCache := network.JSONCache{
		Path: filepath.Join(internal.CachesDir, "quilt", fmt.Sprintf("%s-profile.json", gameVersion)),
	}

	var list FabricVersionList
	err := versionsCache.Read(&list)
	if err != nil {
		err = network.FetchJSON(fmt.Sprintf("https://meta.quiltmc.org/v3/versions/loader/%s", gameVersion), &list)
		if err != nil {
			return FabricMeta{}, fmt.Errorf("retrieve Quilt versions: %w", err)
		}
		versionsCache.Write(list)
	}

	var meta FabricMeta
	if err := profileCache.Read(&meta); err != nil {
		err := network.FetchJSON(fmt.Sprintf("https://meta.quiltmc.org/v3/versions/loader/%s/%s/profile/json", gameVersion, list[0].Loader.Version), &meta)
		if err != nil {
			return FabricMeta{}, fmt.Errorf("retrieve metadata for Quilt version %s: %w", list[0].Loader.Version, err)
		}
		profileCache.Write(meta)
	}

	return meta, nil
}
