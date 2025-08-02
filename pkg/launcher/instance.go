package launcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
	env "github.com/telecter/cmd-launcher/pkg"
)

// An Instance represents a full installation of Minecraft and its information.
type Instance struct {
	Name          string         `toml:"-" json:"-"`
	GameVersion   string         `toml:"game_version" json:"game_version"`
	Loader        Loader         `toml:"mod_loader" json:"mod_loader"`
	LoaderVersion string         `toml:"mod_loader_version,omitempty" json:"mod_loader_version,omitempty"`
	Config        InstanceConfig `toml:"config" json:"config"`
}

// WriteConfig writes the instances configuration to its configuration file.
//
// The Name field is ignored, as it is based on the instance's directory.
func (inst Instance) WriteConfig() error {
	data, _ := toml.Marshal(inst)
	return os.WriteFile(filepath.Join(inst.Dir(), "instance.toml"), data, 0644)
}

// Dir returns the instance's directory
func (inst Instance) Dir() string {
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

// InstanceConfig represents the configurable values of an Instance.
type InstanceConfig struct {
	WindowResolution struct {
		Width  int `toml:"width" json:"width"`
		Height int `toml:"height" json:"height"`
	} `toml:"resolution" json:"resolution"                comment:"Game window resolution"`
	Java      string `toml:"java" json:"java"             comment:"Path to a Java executable. If blank, a Mojang-provided JVM will be downloaded."`
	JavaArgs  string `toml:"java_args" json:"java_args"   comment:"Extra arguments to pass to the JVM"`
	CustomJar string `toml:"custom_jar" json:"custom_jar" comment:"Path to a custom JAR to use instead of the normal Minecraft client"`
	MinMemory int    `toml:"min_memory" json:"min_memory" comment:"Minimum game memory, in MB"`
	MaxMemory int    `toml:"max_memory" json:"max_memory" comment:"Maximum game memory, in MB"`
}

// InstanceOptions are options used to designate an instance's version and other parameters on creation.
type InstanceOptions struct {
	Name          string
	GameVersion   string
	Loader        Loader
	LoaderVersion string

	Config InstanceConfig
}

// CreateInstance creates a new instance with the specified options.
func CreateInstance(options InstanceOptions) (Instance, error) {
	if DoesInstanceExist(options.Name) {
		return Instance{}, fmt.Errorf("instance already exists")
	}

	version, err := fetchVersion(options.Loader, options.GameVersion, options.LoaderVersion)
	if err != nil {
		return Instance{}, err
	}

	inst := Instance{
		Name:          options.Name,
		GameVersion:   version.ID,
		Loader:        options.Loader,
		LoaderVersion: version.LoaderID,
		Config:        options.Config,
	}

	if err := os.MkdirAll(inst.Dir(), 0755); err != nil {
		return Instance{}, fmt.Errorf("create instance directory: %w", err)
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

// FetchInstance retrieves the instance with the specified ID.
func FetchInstance(id string) (Instance, error) {
	if !DoesInstanceExist(id) {
		return Instance{}, fmt.Errorf("instance does not exist")
	}

	dir := filepath.Join(env.InstancesDir, id)

	unmarshaler := toml.Unmarshal
	var data []byte
	var err error

	data, err = os.ReadFile(filepath.Join(dir, "instance.toml"))
	if errors.Is(err, os.ErrNotExist) {
		data, err = os.ReadFile(filepath.Join(dir, "instance.json"))
		if errors.Is(err, os.ErrNotExist) {
			return Instance{}, fmt.Errorf("instance configuration missing")
		} else if err != nil {
			return Instance{}, fmt.Errorf("read instance configuration (JSON): %w", err)
		}
		unmarshaler = json.Unmarshal
	} else if err != nil {
		return Instance{}, fmt.Errorf("read instance configuration: %w", err)
	}

	var inst Instance
	if err := unmarshaler(data, &inst); err != nil {
		return Instance{}, fmt.Errorf("parse instance configuration: %w", err)
	}

	inst.Name = id

	// If instance is using JSON config, migrate it to TOML. Also resets formatting of configuration if changed.
	inst.WriteConfig()
	return inst, nil
}

// FetchAllInstances retrieves all valid instances within env.InstancesDir.
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

// DoesInstanceExist reports whether an instance with the specified ID exists.
func DoesInstanceExist(id string) bool {
	_, err := os.Stat(filepath.Join(env.InstancesDir, id))
	return err == nil
}
