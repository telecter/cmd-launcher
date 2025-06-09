package launcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/telecter/cmd-launcher/internal/meta"
	env "github.com/telecter/cmd-launcher/pkg"
)

// An Instance represents a full installation of Minecraft and its information.
type Instance struct {
	Name          string           `json:"-"`
	Version       meta.VersionMeta `json:"-"`
	GameVersion   string           `json:"game_version"`
	Loader        Loader           `json:"mod_loader"`
	LoaderVersion string           `json:"mod_loader_version,omitempty"`
	Config        InstanceConfig   `json:"config"`
}

// WriteConfig marshals inst and writes it to the instance configuration file. This is used to save the instance configuration.
//
// The Name field is ignored, as it is based on the instance's directory.
func (inst *Instance) WriteConfig() error {
	data, _ := json.MarshalIndent(*inst, "", "    ")
	return os.WriteFile(filepath.Join(inst.Dir(), "instance.json"), data, 0644)
}

func (inst *Instance) Dir() string {
	return filepath.Join(env.InstancesDir, inst.Name)
}

// Rename renames instance to the specified new name
func (inst *Instance) Rename(new string) error {
	if err := os.Rename(inst.Dir(), filepath.Join(env.InstancesDir, new)); err != nil {
		return err
	}
	inst.Name = new
	return nil
}

type InstanceOptions struct {
	Name          string
	GameVersion   string
	Loader        Loader
	LoaderVersion string
}

// InstanceConfig represents the configurable values of an Instance.
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
	MinMemory: 512,
	MaxMemory: 4096,
}

// CreateInstance creates a new instance with the specified options.
func CreateInstance(options InstanceOptions) (Instance, error) {
	if IsInstanceExist(options.Name) {
		return Instance{}, fmt.Errorf("instance already exists")
	}

	versionMeta, err := meta.FetchVersionMeta(options.GameVersion)
	if err != nil {
		return Instance{}, fmt.Errorf("fetch version meta: %w", err)
	}

	loaderMeta, err := fetchLoaderMeta(versionMeta, options.Loader, options.LoaderVersion)
	if err != nil {
		return Instance{}, fmt.Errorf("fetch mod loader meta: %w", err)
	}
	versionMeta.LoaderID = loaderMeta.ID

	versionMeta = meta.MergeVersionMeta(versionMeta, loaderMeta)

	inst := Instance{
		Name:          options.Name,
		GameVersion:   versionMeta.ID,
		Version:       versionMeta,
		Loader:        options.Loader,
		LoaderVersion: versionMeta.LoaderID,
		Config:        defaultConfig,
	}

	if err := os.MkdirAll(inst.Dir(), 0755); err != nil {
		return Instance{}, fmt.Errorf("create instance directory: %w", err)
	}

	java, err := exec.LookPath("java")
	if err == nil {
		inst.Config.Java = java
	}

	if err := inst.WriteConfig(); err != nil {
		return Instance{}, fmt.Errorf("write instance configuration: %w", err)
	}

	return inst, nil
}

// RemoveInstance removes the instance with the specified ID.
func RemoveInstance(id string) error {
	inst, err := FetchInstance(id)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(inst.Dir()); err != nil {
		return fmt.Errorf("remove instance directory: %w", err)
	}
	return nil
}

// GetInstance retrieves the instance with the specified ID.
func FetchInstance(id string) (Instance, error) {
	dir := filepath.Join(env.InstancesDir, id)
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
	inst.Name = id

	versionMeta, err := meta.FetchVersionMeta(inst.GameVersion)
	if err != nil {
		return Instance{}, fmt.Errorf("fetch version meta: %w", err)
	}
	versionMeta.LoaderID = inst.LoaderVersion

	loaderMeta, err := fetchLoaderMeta(versionMeta, inst.Loader, inst.LoaderVersion)
	if err != nil {
		return Instance{}, fmt.Errorf("fetch mod loader meta: %w", err)
	}
	versionMeta = meta.MergeVersionMeta(versionMeta, loaderMeta)

	inst.Version = versionMeta
	return inst, nil
}

// AllInstances retrieves all valid instances within env.InstancesDir.
func FetchAllInstances() ([]Instance, error) {
	entries, err := os.ReadDir(env.InstancesDir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read instances directory: %w", err)
	}
	var insts []Instance
	for _, entry := range entries {
		if entry.IsDir() {
			inst, err := FetchInstance(entry.Name())
			if err != nil {
				continue
			}
			insts = append(insts, inst)
		}
	}
	return insts, nil
}

// IsInstanceExist reports whether an instance with the specified ID exists.
func IsInstanceExist(id string) bool {
	_, err := FetchInstance(id)
	return err == nil
}
