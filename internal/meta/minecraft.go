package meta

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/telecter/cmd-launcher/internal"
	"github.com/telecter/cmd-launcher/internal/network"
)

type VersionManifest struct {
	Latest struct {
		Release  string `json:"release"`
		Snapshot string `json:"snapshot"`
	} `json:"latest"`
	Versions []struct {
		ID              string    `json:"id"`
		Type            string    `json:"type"`
		URL             string    `json:"url"`
		Time            time.Time `json:"time"`
		ReleaseTime     time.Time `json:"releaseTime"`
		Sha1            string    `json:"sha1"`
		ComplianceLevel int       `json:"complianceLevel"`
	} `json:"versions"`
}

type VersionMeta struct {
	Arguments struct {
		Game []any `json:"game"`
		Jvm  []any `json:"jvm"`
	} `json:"arguments"`
	AssetIndex struct {
		ID        string `json:"id"`
		Sha1      string `json:"sha1"`
		Size      int    `json:"size"`
		TotalSize int    `json:"totalSize"`
		URL       string `json:"url"`
	} `json:"assetIndex"`
	Assets          string `json:"assets"`
	ComplianceLevel int    `json:"complianceLevel"`
	Downloads       struct {
		Client struct {
			Sha1 string `json:"sha1"`
			Size int    `json:"size"`
			URL  string `json:"url"`
		} `json:"client"`
		ClientMappings struct {
			Sha1 string `json:"sha1"`
			Size int    `json:"size"`
			URL  string `json:"url"`
		} `json:"client_mappings"`
		Server struct {
			Sha1 string `json:"sha1"`
			Size int    `json:"size"`
			URL  string `json:"url"`
		} `json:"server"`
		ServerMappings struct {
			Sha1 string `json:"sha1"`
			Size int    `json:"size"`
			URL  string `json:"url"`
		} `json:"server_mappings"`
	} `json:"downloads"`
	ID          string `json:"id"`
	JavaVersion struct {
		Component    string `json:"component"`
		MajorVersion int    `json:"majorVersion"`
	} `json:"javaVersion"`
	Libraries []Library `json:"libraries"`
	Logging   struct {
		Client struct {
			Argument string `json:"argument"`
			File     struct {
				ID   string `json:"id"`
				Sha1 string `json:"sha1"`
				Size int    `json:"size"`
				URL  string `json:"url"`
			} `json:"file"`
			Type string `json:"type"`
		} `json:"client"`
	} `json:"logging"`
	MainClass              string    `json:"mainClass"`
	MinimumLauncherVersion int       `json:"minimumLauncherVersion"`
	ReleaseTime            time.Time `json:"releaseTime"`
	Time                   time.Time `json:"time"`
	Type                   string    `json:"type"`
}
type Artifact struct {
	Path string `json:"path"`
	Sha1 string `json:"sha1"`
	Size int    `json:"size"`
	URL  string `json:"url"`
}
type Library struct {
	Downloads struct {
		Artifact Artifact `json:"artifact"`
	} `json:"downloads"`
	Name    string `json:"name"`
	Natives struct {
		Linux   string `json:"linux"`
		MacOS   string `json:"macos"`
		Windows string `json:"windows"`
	}
	// these fields are present in Fabric libraries that don't contain a 'downloads' field
	URL   string `json:"url"`
	Sha1  string `json:"sha1"`
	Size  int    `json:"size"`
	Rules []struct {
		Action string `json:"action"`
		Os     struct {
			Name string `json:"name"`
		} `json:"os"`
	} `json:"rules,omitempty"`
}
type AssetIndex struct {
	Objects map[string]AssetObject `json:"objects"`
}
type AssetObject struct {
	Hash string `json:"hash"`
	Size int    `json:"size"`
}

func GetVersionManifest() (VersionManifest, error) {
	cache := network.JSONCache{Path: filepath.Join(internal.CachesDir, "minecraft", "version_manifest.json")}

	var manifest VersionManifest
	if err := cache.Read(&manifest); err != nil {
		if err := network.FetchJSON("https://piston-meta.mojang.com/mc/game/version_manifest_v2.json", &manifest); err != nil {
			return VersionManifest{}, fmt.Errorf("retrieve version manifest: %w", err)
		}
		cache.Write(manifest)
	}
	return manifest, nil
}

func GetVersionMeta(id string) (VersionMeta, error) {
	manifest, err := GetVersionManifest()
	if err != nil {
		return VersionMeta{}, err
	}
	for _, v := range manifest.Versions {
		if v.ID == id {
			cache := network.JSONCache{
				Path: filepath.Join(internal.CachesDir, "minecraft", fmt.Sprintf("%s.json", v.ID)),
			}
			var versionMeta VersionMeta
			if err := cache.Read(&versionMeta); err != nil {
				if err := network.FetchJSON(v.URL, &versionMeta); err != nil {
					return VersionMeta{}, fmt.Errorf("retrieve version metadata: %w", err)
				}
				cache.Write(versionMeta)
			}
			versionMeta.Libraries = append(versionMeta.Libraries, Library{
				Name: "com.mojang:minecraft:" + versionMeta.ID,
				Downloads: struct {
					Artifact Artifact "json:\"artifact\""
				}{
					Artifact: Artifact{
						Path: fmt.Sprintf("com/mojang/minecraft/%s/%s.jar", versionMeta.ID, versionMeta.ID),
						Sha1: versionMeta.Downloads.Client.Sha1,
						Size: versionMeta.Downloads.Client.Size,
						URL:  versionMeta.Downloads.Client.URL,
					},
				}})
			return versionMeta, nil
		}
	}
	return VersionMeta{}, fmt.Errorf("invalid version")
}
