package launcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/telecter/cmd-launcher/internal"
	"github.com/telecter/cmd-launcher/internal/meta"
)

type InstanceOptions struct {
	GameVersion string
	Name        string
	Loader      Loader
}
type Instance struct {
	Dir           string         `json:"-"`
	Name          string         `json:"-"`
	GameVersion   string         `json:"game_version"`
	Loader        Loader         `json:"mod_loader"`
	LoaderVersion string         `json:"mod_loader_version,omitempty"`
	Config        InstanceConfig `json:"config"`
}
type InstanceConfig struct {
	WindowResolution struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"resolution"`
	Java      string `json:"java"`
	MinMemory int    `json:"min_memory"`
	MaxMemory int    `json:"max_memory"`
}

var defaultConfig = InstanceConfig{
	WindowResolution: struct {
		Width  int "json:\"width\""
		Height int "json:\"height\""
	}{
		Width:  1708,
		Height: 960,
	},
	Java:      "/usr/bin/java",
	MinMemory: 512,
	MaxMemory: 4096,
}

func CreateInstance(options InstanceOptions) (Instance, error) {
	if IsInstanceExist(options.Name) {
		return Instance{}, fmt.Errorf("instance already exists")
	}

	if options.GameVersion == "release" || options.GameVersion == "snapshot" {
		manifest, err := meta.GetVersionManifest()
		if err != nil {
			return Instance{}, err
		}
		if options.GameVersion == "release" {
			options.GameVersion = manifest.Latest.Release
		} else if options.GameVersion == "snapshot" {
			options.GameVersion = manifest.Latest.Snapshot
		}
	}

	if _, err := meta.GetVersionMeta(options.GameVersion); err != nil {
		return Instance{}, err
	}

	var loaderVersion string
	if options.Loader == LoaderFabric || options.Loader == LoaderQuilt {
		var fabricLoader meta.FabricLoader
		switch options.Loader {
		case LoaderFabric:
			fabricLoader = meta.FabricLoaderStandard
		case LoaderQuilt:
			fabricLoader = meta.FabricLoaderQuilt
		}
		fabricVersions, err := meta.GetFabricVersions(fabricLoader)
		if err != nil {
			return Instance{}, fmt.Errorf("retrieve %s versions: %w", fabricLoader.String(), err)
		}
		loaderVersion = fabricVersions[0].Version
	}

	dir := filepath.Join(internal.InstancesDir, options.Name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return Instance{}, fmt.Errorf("create instance directory: %w", err)
	}

	inst := Instance{
		Dir:           dir,
		Name:          options.Name,
		GameVersion:   options.GameVersion,
		Loader:        options.Loader,
		LoaderVersion: loaderVersion,
		Config:        defaultConfig,
	}

	data, _ := json.MarshalIndent(inst, "", "    ")
	if err := os.WriteFile(filepath.Join(dir, "instance.json"), data, 0644); err != nil {
		return Instance{}, fmt.Errorf("write instant metadata: %w", err)
	}

	return inst, nil
}

func RemoveInstance(id string) error {
	inst, err := GetInstance(id)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(inst.Dir); err != nil {
		return fmt.Errorf("remove instance directory: %w", err)
	}
	return nil
}

func GetInstance(id string) (Instance, error) {
	dir := filepath.Join(internal.InstancesDir, id)
	data, err := os.ReadFile(filepath.Join(dir, "instance.json"))
	if errors.Is(err, os.ErrNotExist) {
		return Instance{}, fmt.Errorf("instance does not exist")
	} else if err != nil {
		return Instance{}, fmt.Errorf("read instance metadata: %w", err)
	}
	var inst Instance
	if err := json.Unmarshal(data, &inst); err != nil {
		return Instance{}, fmt.Errorf("parse instance metadata: %w", err)
	}
	inst.Dir = dir
	inst.Name = id
	return inst, nil
}

func GetAllInstances() ([]Instance, error) {
	entries, err := os.ReadDir(internal.InstancesDir)
	if errors.Is(err, os.ErrNotExist) {
		return []Instance{}, nil
	}
	if err != nil {
		return []Instance{}, fmt.Errorf("read instances directory: %w", err)
	}
	var insts []Instance
	for _, entry := range entries {
		if entry.IsDir() {
			inst, err := GetInstance(entry.Name())
			if err != nil {
				continue
			}
			insts = append(insts, inst)
		}
	}
	return insts, nil
}

func IsInstanceExist(id string) bool {
	_, err := GetInstance(id)
	return err == nil
}
