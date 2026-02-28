package meta

import (
	"fmt"
)

// Loader represents a game mod loader.
type Loader string

const (
	LoaderVanilla  Loader = "vanilla"
	LoaderFabric   Loader = "fabric"
	LoaderQuilt    Loader = "quilt"
	LoaderNeoForge Loader = "neoforge"
	LoaderForge    Loader = "forge"
)

// FetchAllVersionMeta returns a VersionMeta containing both information for the base game, and specified mod loader.
func FetchAllVersionMeta(loader Loader, gameVersion string, loaderVersion string) (VersionMeta, error) {
	var loaderMeta VersionMeta
	var err error

	version, err := FetchVersionMeta(gameVersion)
	if err != nil {
		return VersionMeta{}, fmt.Errorf("retrieve version metadata: %w", err)
	}

	switch loader {
	case LoaderFabric, LoaderQuilt:
		api := Fabric
		if loader == LoaderQuilt {
			api = Quilt
		}
		loaderMeta, err = api.FetchMeta(version.ID, loaderVersion)
		if err != nil {
			return VersionMeta{}, fmt.Errorf("retrieve Fabric/Quilt metadata: %w", err)
		}
	case LoaderNeoForge:
		if loaderVersion == "latest" {
			loaderVersion, err = FetchNeoforgeVersion(version.ID)
			if err != nil {
				return VersionMeta{}, fmt.Errorf("retrieve NeoForge version: %w", err)
			}
		}
		loaderMeta, _, err = Neoforge.FetchMeta(loaderVersion)
		if err != nil {
			return VersionMeta{}, fmt.Errorf("retrieve NeoForge metadata: %w", err)
		}
	case LoaderForge:
		if loaderVersion == "latest" {
			loaderVersion, err = FetchForgeVersion(version.ID)
			if err != nil {
				return VersionMeta{}, fmt.Errorf("retrieve Forge version: %w", err)
			}
		}
		loaderMeta, _, err = Forge.FetchMeta(loaderVersion)
		if err != nil {
			return VersionMeta{}, fmt.Errorf("retrieve Forge metadata: %w", err)
		}
	}

	return MergeVersionMeta(version, loaderMeta), nil
}
