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
	Loader      string
}
type Instance struct {
	Dir         string         `json:"dir"`
	GameVersion string         `json:"game_version"`
	Name        string         `json:"name"`
	Loader      string         `json:"mod_loader"`
	Config      InstanceConfig `json:"config"`
}
type InstanceConfig struct {
	WindowResolution struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"window_resolution"`
	JavaExecutablePath string `json:"java_location"`
	MinMemory          int    `json:"min_memory"`
	MaxMemory          int    `json:"max_memory"`
}

var defaultConfig = InstanceConfig{
	WindowResolution: struct {
		Width  int "json:\"width\""
		Height int "json:\"height\""
	}{
		Width:  1708,
		Height: 960,
	},
	JavaExecutablePath: "/usr/bin/java",
	MinMemory:          512,
	MaxMemory:          4096,
}

func CreateInstance(options InstanceOptions) (Instance, error) {
	if IsInstanceExist(options.Name) {
		return Instance{}, fmt.Errorf("instance already exists")
	}
	if options.Loader != LoaderFabric && options.Loader != LoaderVanilla {
		return Instance{}, fmt.Errorf("invalid mod loader")
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

	dir := filepath.Join(internal.InstancesDir, options.Name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return Instance{}, fmt.Errorf("failed to create instance directory: %w", err)
	}

	inst := Instance{
		Dir:         dir,
		GameVersion: options.GameVersion,
		Loader:      options.Loader,
		Name:        options.Name,
		Config:      defaultConfig,
	}

	data, _ := json.Marshal(inst)
	os.WriteFile(filepath.Join(dir, "instance.json"), data, 0644)

	return inst, nil
}
func DeleteInstance(id string) error {
	inst, err := GetInstance(id)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(inst.Dir); err != nil {
		return fmt.Errorf("failed to remove instance directory: %w", err)
	}
	return nil
}

func GetInstance(id string) (Instance, error) {
	dir := filepath.Join(internal.InstancesDir, id)
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		return Instance{}, errors.New("instance does not exist")
	}
	data, err := os.ReadFile(filepath.Join(dir, "instance.json"))
	if err != nil {
		return Instance{}, fmt.Errorf("failed to read instance metadata: %w", err)
	}
	var inst Instance
	if err := json.Unmarshal(data, &inst); err != nil {
		return Instance{}, fmt.Errorf("instance metadata is invalid: %w", err)
	}
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
	if _, err := GetInstance(id); err != nil {
		return false
	}
	return true
}
