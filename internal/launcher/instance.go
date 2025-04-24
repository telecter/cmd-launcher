package launcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/telecter/cmd-launcher/internal/env"
	"github.com/telecter/cmd-launcher/internal/network"
	"github.com/telecter/cmd-launcher/internal/network/api"
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

func (instance *Instance) Start(classpath []string, options LaunchOptions) error {
	meta, err := instance.GetVersionMeta()
	if err != nil {
		return fmt.Errorf("failed to get version metadata: %w", err)
	}

	if err := instance.DownloadClient(); err != nil {
		return fmt.Errorf("failed to download client: %w", err)
	}

	classpath = append(classpath, filepath.Join(instance.Dir, instance.GameVersion+".jar"))

	jvmArgs := []string{"-cp", strings.Join(classpath, ":")}
	if runtime.GOOS == "darwin" {
		jvmArgs = append(jvmArgs, "-XstartOnFirstThread")
	}
	if instance.Config.MinMemory != 0 {
		jvmArgs = append(jvmArgs, fmt.Sprintf("-Xms%dm", instance.Config.MinMemory))
	}
	if instance.Config.MaxMemory != 0 {
		jvmArgs = append(jvmArgs, fmt.Sprintf("-Xmx%dm", instance.Config.MaxMemory))
	}
	if instance.ModLoader == "fabric" {
		fabricMeta, err := instance.GetFabricMeta()
		if err != nil {
			return err
		}
		jvmArgs = append(jvmArgs, fabricMeta.Arguments.Jvm...)
		jvmArgs = append(jvmArgs, fabricMeta.MainClass)
	} else {
		jvmArgs = append(jvmArgs, meta.MainClass)
	}

	gameArgs := []string{
		"--username", options.LoginData.Username,
		"--accessToken", options.LoginData.Token,
		"--gameDir", instance.Dir,
		"--assetsDir", filepath.Join(env.RootDir, "assets"),
		"--assetIndex", meta.AssetIndex.ID,
		"--version", instance.GameVersion,
		"--versionType", meta.Type,
		"--width", strconv.Itoa(instance.Config.WindowResolution[0]),
		"--height", strconv.Itoa(instance.Config.WindowResolution[1]),
	}
	if options.QuickPlayServer != "" {
		gameArgs = append(gameArgs, "--quickPlayMultiplayer", options.QuickPlayServer)
	}
	if options.LoginData.UUID != "" {
		gameArgs = append(gameArgs, "--uuid", options.LoginData.UUID)
	}
	os.Chdir(instance.Dir)
	return run(instance.Config.JavaExecutablePath, append(jvmArgs, gameArgs...))
}

func CreateInstance(options InstanceOptions) (Instance, error) {
	if options.ModLoader != "" && options.ModLoader != "fabric" {
		return Instance{}, fmt.Errorf("invalid mod loader")
	}

	if IsInstanceExist(options.Name) {
		return Instance{}, fmt.Errorf("instance already exists")
	}

	if options.GameVersion == "release" {
		id, _ := api.GetLatestRelease()
		options.GameVersion = id
	} else if options.GameVersion == "snapshot" {
		id, _ := api.GetLatestSnapshot()
		options.GameVersion = id
	}

	if _, err := api.GetVersionMeta(options.GameVersion); err != nil {
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

func (instance *Instance) GetVersionMeta() (api.VersionMeta, error) {
	var meta api.VersionMeta
	if data, err := os.ReadFile(filepath.Join(instance.Dir, instance.GameVersion+".json")); err == nil {
		json.Unmarshal(data, &meta)
	} else {
		meta, err = api.GetVersionMeta(instance.GameVersion)
		if err != nil {
			return api.VersionMeta{}, err
		}
		json, _ := json.Marshal(meta)
		os.WriteFile(filepath.Join(instance.Dir, instance.GameVersion+".json"), json, 0644)
	}
	return meta, nil
}

func (instance *Instance) GetFabricMeta() (api.FabricMeta, error) {
	var meta api.FabricMeta
	if data, err := os.ReadFile(filepath.Join(instance.Dir, "fabric.json")); err == nil {
		json.Unmarshal(data, &meta)
	} else {
		meta, err = api.GetLoaderMeta(instance.GameVersion)
		if err != nil {
			return api.FabricMeta{}, err
		}
		data, _ := json.Marshal(meta)
		os.WriteFile(filepath.Join(instance.Dir, "fabric.json"), data, 0644)
	}
	return meta, nil
}

func (instance *Instance) DownloadClient() error {
	meta, err := instance.GetVersionMeta()
	if err != nil {
		return fmt.Errorf("failed to get version metadata: %w", err)
	}

	if err := network.DownloadFile(meta.Downloads.Client.URL, filepath.Join(instance.Dir, instance.GameVersion+".jar")); err != nil {
		return fmt.Errorf("failed to download client: %s", err)
	}
	return nil
}
