package meta

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/telecter/cmd-launcher/internal/network"
	env "github.com/telecter/cmd-launcher/pkg"
)

// A VersionManifest is a list of all Minecraft versions.
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

// A VersionMeta is metadata of the libraries, assets, and other data needed to start a Minecraft version.
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
	Libraries []MojangLibrary `json:"libraries"`
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

// Client creates a library from the client JAR download of versionMeta.
func (versionMeta VersionMeta) Client() MojangLibrary {
	specifier, _ := NewLibrarySpecifier("com.mojang:minecraft:" + versionMeta.ID)
	return MojangLibrary{
		Name: specifier,
		Downloads: struct {
			Artifact    Artifact            "json:\"artifact\""
			Classifiers map[string]Artifact "json:\"classifiers\""
		}{
			Artifact: Artifact{
				Path: fmt.Sprintf("com/mojang/minecraft/%s/%s.jar", versionMeta.ID, versionMeta.ID),
				Sha1: versionMeta.Downloads.Client.Sha1,
				Size: versionMeta.Downloads.Client.Size,
				URL:  versionMeta.Downloads.Client.URL,
			},
		},
	}
}

// An Artifact represents a library JAR file that can be downloaded
type Artifact struct {
	Path string `json:"path"`
	Sha1 string `json:"sha1"`
	Size int    `json:"size"`
	URL  string `json:"url"`
}

func (artifact Artifact) RuntimePath() string {
	return filepath.Join(env.LibrariesDir, artifact.Path)
}
func (artifact Artifact) IsDownloaded() bool {
	data, err := os.ReadFile(artifact.RuntimePath())
	if err != nil {
		return false
	}
	// if no checksum is present, still count the artifact as installed as long as the file exists
	if artifact.Sha1 == "" {
		return true
	}

	sum := sha1.Sum(data)
	return artifact.Sha1 == hex.EncodeToString(sum[:])
}
func (artifact Artifact) DownloadEntry() network.DownloadEntry {
	return network.DownloadEntry{
		URL:      artifact.URL,
		Filename: artifact.RuntimePath(),
	}
}

// A Library represents metadata of a game library and its artifact(s).
type Library interface {
	Artifact() Artifact
	ShouldInstall() bool
	Specifier() LibrarySpecifier
}

type MojangLibrary struct {
	Name      LibrarySpecifier `json:"name"`
	Downloads struct {
		Artifact    Artifact            `json:"artifact"`
		Classifiers map[string]Artifact `json:"classifiers"`
	} `json:"downloads"` // mojang
	Rules []struct {
		Action string `json:"action"`
		Os     struct {
			Name string `json:"name"`
		} `json:"os"`
	} `json:"rules,omitempty"` // mojang
	Natives map[string]string `json:"natives,omitempty"` // old mojang
}

func (library MojangLibrary) Artifact() Artifact {
	return library.Downloads.Artifact
}
func (library MojangLibrary) Classifiers() (natives []Library) {
	var classifiers []string
	for os, native := range library.Natives {
		os = strings.ReplaceAll(os, "osx", "darwin")
		if os == runtime.GOOS {
			classifiers = append(classifiers, native)
		}
	}
	for _, classifier := range classifiers {
		artifact := library.Downloads.Classifiers[classifier]
		if artifact.URL == "" {
			continue
		}
		specifier := library.Name
		specifier.Classifier = classifier
		natives = append(natives, BaseLibrary{
			LibraryArtifact: artifact,
			Name:            specifier,
		})
	}
	return natives
}

func (library MojangLibrary) Specifier() LibrarySpecifier {
	return library.Name
}

// ShouldInstall reports whether the Rules field on library allows library to be installed.
func (library MojangLibrary) ShouldInstall() bool {
	if len(library.Rules) > 0 {
		rule := library.Rules[0]
		os := strings.ReplaceAll(rule.Os.Name, "osx", "darwin")
		return rule.Action == "allow" && (os == runtime.GOOS || os == "")
	}
	return true
}

type BaseLibrary struct {
	LibraryArtifact Artifact
	Name            LibrarySpecifier
}

func (library BaseLibrary) Artifact() Artifact {
	return library.LibraryArtifact
}
func (library BaseLibrary) Specifier() LibrarySpecifier {
	return library.Name
}
func (BaseLibrary) ShouldInstall() bool {
	return true
}

// An AssetIndex contains a map of asset objects and their names.
type AssetIndex struct {
	Objects map[string]AssetObject `json:"objects"`
}

// An AssetObject is a reference to an game asset which can be downloaded.
type AssetObject struct {
	Hash string `json:"hash"`
	Size int    `json:"size"`
}

// DownloadEntry returns a DownloadEntry to fetch asset.
func (object AssetObject) DownloadEntry() network.DownloadEntry {
	return network.DownloadEntry{
		URL:      fmt.Sprintf(MINECRAFT_RESOURCES_URL, object.Hash[:2], object.Hash),
		Filename: filepath.Join(env.AssetsDir, "objects", object.Hash[:2], object.Hash),
	}
}

// IsDownloaded reports whether asset exists and has a valid checksum.
func (object AssetObject) IsDownloaded() bool {
	data, err := os.ReadFile(object.DownloadEntry().Filename)
	if err != nil {
		return false
	}
	sum := sha1.Sum(data)
	return object.Hash == hex.EncodeToString(sum[:])
}

const VERSION_MANIFEST_URL = "https://piston-meta.mojang.com/mc/game/version_manifest_v2.json"
const MINECRAFT_RESOURCES_URL = "https://resources.download.minecraft.net/%s/%s"

// GetVersionManifest retrieves the Mojang version manifest which lists all game versions.
func GetVersionManifest() (VersionManifest, error) {
	cache := network.JSONCache[VersionManifest]{
		Path: filepath.Join(env.CachesDir, "minecraft", "version_manifest.json"),
		URL:  VERSION_MANIFEST_URL,
	}

	var manifest VersionManifest

	if err := cache.UpdateAndRead(&manifest); err != nil {
		if err := cache.Read(&manifest); err != nil {
			return VersionManifest{}, err
		}
	}
	return manifest, nil
}

// GetVersionMeta retrieves the version metadata for a specified version from the version manifest.
func GetVersionMeta(id string) (VersionMeta, error) {
	manifest, err := GetVersionManifest()
	if err != nil {
		return VersionMeta{}, fmt.Errorf("retrieve version manifest: %w", err)
	}
	for _, v := range manifest.Versions {
		if v.ID == id {
			cache := network.JSONCache[VersionMeta]{
				Path: filepath.Join(env.CachesDir, "minecraft", v.ID+".json"),
				URL:  v.URL,
			}
			download := true

			var versionMeta VersionMeta
			if err := cache.Read(&versionMeta); err == nil {
				sum, _ := cache.Sha1()
				if sum == v.Sha1 {
					download = false
				}
			}
			if download {
				if err := cache.UpdateAndRead(&versionMeta); err != nil {
					return VersionMeta{}, err
				}
			}
			return versionMeta, nil
		}
	}
	return VersionMeta{}, fmt.Errorf("invalid version")
}

// DownloadAssetIndex retrieves the asset index for the specified version.
func DownloadAssetIndex(versionMeta VersionMeta) (AssetIndex, error) {
	cache := network.JSONCache[AssetIndex]{
		Path: filepath.Join(env.AssetsDir, "indexes", versionMeta.AssetIndex.ID+".json"),
		URL:  versionMeta.AssetIndex.URL,
	}
	download := true

	var assetIndex AssetIndex
	if err := cache.Read(&assetIndex); err == nil {
		sum, _ := cache.Sha1()
		if sum == versionMeta.AssetIndex.Sha1 {
			download = false
		}
	}
	if download {
		if err := cache.UpdateAndRead(&assetIndex); err != nil {
			return AssetIndex{}, err
		}
	}

	return assetIndex, nil
}
