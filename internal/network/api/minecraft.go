package api

import (
	"fmt"
	"time"

	"github.com/telecter/cmd-launcher/internal/network"
)

type VersionManifest struct {
	Latest struct {
		Release  string `json:"release"`
		Snapshot string `json:"snapshot"`
	} `json:"latest"`
	Versions []Version `json:"versions"`
}
type Version struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	URL         string    `json:"url"`
	Time        time.Time `json:"time"`
	ReleaseTime time.Time `json:"releaseTime"`
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
	URL   string `json:"url"`
	Rules []struct {
		Action string `json:"action"`
		Os     struct {
			Name string `json:"name"`
		} `json:"os"`
	} `json:"rules,omitempty"`
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
type AssetIndex struct {
	Objects map[string]struct {
		Hash string
		Size int
	}
}

func GetVersionManifest() (VersionManifest, error) {
	var manifest VersionManifest
	err := network.FetchJSONData("https://launchermeta.mojang.com/mc/game/version_manifest.json", &manifest)
	if err != nil {
		return VersionManifest{}, fmt.Errorf("failed to retrieve version manifest: %w", err)
	}
	return manifest, nil
}

func GetLatestRelease() (string, error) {
	manifest, err := GetVersionManifest()
	if err != nil {
		return "", err
	}
	return manifest.Latest.Release, err
}

func GetVersionMeta(id string) (VersionMeta, error) {
	manifest, err := GetVersionManifest()
	if err != nil {
		return VersionMeta{}, err
	}

	for _, v := range manifest.Versions {
		if v.ID == id {
			var meta VersionMeta
			err := network.FetchJSONData(v.URL, &meta)
			if err != nil {
				return VersionMeta{}, fmt.Errorf("failed to retrieve version metadata: %w", err)
			}
			return meta, nil
		}
	}
	return VersionMeta{}, fmt.Errorf("invalid version")
}
