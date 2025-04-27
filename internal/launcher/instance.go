package launcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/telecter/cmd-launcher/internal/env"
	"github.com/telecter/cmd-launcher/internal/meta"
)

type InstanceOptions struct {
	GameVersion string
	Name        string
	ModLoader   string
}
type Instance struct {
	Dir         string         `json:"dir"`
	GameVersion string         `json:"game_version"`
	Name        string         `json:"name"`
	ModLoader   string         `json:"mod_loader"`
	Config      InstanceConfig `json:"config"`
}
type InstanceConfig struct {
	WindowResolution   [2]int `json:"window_resolution"`
	JavaExecutablePath string `json:"java_location"`
	MinMemory          int    `json:"min_memory"`
	MaxMemory          int    `json:"max_memory"`
}

var defaultInstanceConfig = InstanceConfig{
	WindowResolution:   [2]int{1708, 960},
	JavaExecutablePath: "/usr/bin/java",
	MinMemory:          512,
	MaxMemory:          4096,
}

func CreateInstance(options InstanceOptions) (Instance, error) {
	if options.ModLoader != "" && options.ModLoader != "fabric" {
		return Instance{}, fmt.Errorf("invalid mod loader")
	}

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

	dir := filepath.Join(env.InstancesDir, options.Name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return Instance{}, fmt.Errorf("failed to create instance directory: %w", err)
	}

	instance := Instance{
		Dir:         dir,
		GameVersion: options.GameVersion,
		ModLoader:   options.ModLoader,
		Name:        options.Name,
		Config:      defaultInstanceConfig,
	}

	data, _ := json.Marshal(instance)
	os.WriteFile(filepath.Join(dir, "instance.json"), data, 0644)

	return instance, nil
}
func DeleteInstance(id string) error {
	instance, err := GetInstance(id)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(instance.Dir); err != nil {
		return fmt.Errorf("failed to remove instance directory: %w", err)
	}
	return nil
}

func GetInstance(id string) (Instance, error) {
	dir := filepath.Join(env.InstancesDir, id)
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		return Instance{}, errors.New("instance does not exist")
	}
	data, err := os.ReadFile(filepath.Join(dir, "instance.json"))
	if err != nil {
		return Instance{}, fmt.Errorf("failed to read instance metadata: %w", err)
	}
	var instance Instance
	if err := json.Unmarshal(data, &instance); err != nil {
		return Instance{}, fmt.Errorf("instance metadata is invalid: %w", err)
	}
	return instance, nil
}

func GetAllInstances() ([]Instance, error) {
	entries, err := os.ReadDir(env.InstancesDir)
	if err != nil {
		return []Instance{}, fmt.Errorf("failed to read instances directory: %w", err)
	}
	var instances []Instance
	for _, entry := range entries {
		if entry.IsDir() {
			instance, err := GetInstance(entry.Name())
			if err != nil {
				continue
			}
			instances = append(instances, instance)
		}
	}
	return instances, nil
}

func IsInstanceExist(id string) bool {
	if _, err := GetInstance(id); err != nil {
		return false
	}
	return true
}
