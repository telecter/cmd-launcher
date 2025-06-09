// Package launcher provides the necessary functions to start the game.
package launcher

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/telecter/cmd-launcher/internal/meta"
	"github.com/telecter/cmd-launcher/internal/network"
	env "github.com/telecter/cmd-launcher/pkg"
	"github.com/telecter/cmd-launcher/pkg/auth"
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

func (loader Loader) String() string {
	return string(loader)
}

// EnvOptions represents configuration options when preparing an instance to be launched.
type EnvOptions struct {
	Session            auth.Session
	Config             InstanceConfig
	QuickPlayServer    string
	Demo               bool
	DisableMultiplayer bool
	DisableChat        bool

	skipAssets    bool
	skipLibraries bool
}

// An EventWatcher is a controller that handles game preparation events.
type EventWatcher interface {
	Handle(event any)
}

// MetadataResolvedEvent is called when all metadata has been retrieved
type MetadataResolvedEvent struct{}

// LibrariesResolvedEvent is called when all game libraries have been identified and filtered.
type LibrariesResolvedEvent struct {
	Libraries int
}

// AssetsResolvedEvent is called when all game assets have been identified and filtered.
type AssetsResolvedEvent struct {
	Assets int
}

// DownloadingEvent is called when a download has progressed.
type DownloadingEvent struct {
	Completed int
	Total     int
}

// A Runner is a controller which manages the starting of the game.
type Runner interface {
	Run(cmd *exec.Cmd) error
}

// An ConsoleRunner is an implementation of Runner which logs game output to the console.
type ConsoleRunner struct{}

func (ConsoleRunner) Run(cmd *exec.Cmd) error {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// A LaunchEnvironment represents the data needed to start the game.
type LaunchEnvironment struct {
	GameDir   string
	JavaPath  string
	MainClass string
	Classpath []string
	JavaArgs  []string
	GameArgs  []string
}

// Launch starts a LaunchEnvironment with the specified runner.
//
// The Java executable is checked and the classpath and command arguments are finalized.
func Launch(launchEnv LaunchEnvironment, runner Runner) error {
	info, err := os.Stat(launchEnv.JavaPath)
	if err != nil {
		return fmt.Errorf("java executable does not exist")
	}
	if info.Mode()&0111 == 0 || info.IsDir() {
		return fmt.Errorf("java binary is not executable")
	}

	javaArgs := append(launchEnv.JavaArgs, "-cp", strings.Join(launchEnv.Classpath, string(os.PathListSeparator)), launchEnv.MainClass)
	cmd := exec.Command(launchEnv.JavaPath, append(javaArgs, launchEnv.GameArgs...)...)
	cmd.Dir = launchEnv.GameDir
	return runner.Run(cmd)
}

// Prepare prepares the instance to be launched, returning a LaunchEnvironment, with the provided options and sends events to watcher.
func Prepare(inst Instance, options EnvOptions, watcher EventWatcher) (LaunchEnvironment, error) {
	launchEnv := LaunchEnvironment{
		JavaPath: options.Config.Java,
		GameDir:  inst.Dir(),
	}

	version, err := meta.FetchVersionMeta(inst.GameVersion)
	if err != nil {
		return LaunchEnvironment{}, fmt.Errorf("retrieve version metadata: %w", err)
	}
	loaderMeta, err := fetchLoaderMeta(inst.Loader, version.ID, inst.LoaderVersion)
	if err != nil {
		return LaunchEnvironment{}, fmt.Errorf("retrieve loader metadata: %w", err)
	}
	version = meta.MergeVersionMeta(version, loaderMeta)

	launchEnv.MainClass = version.MainClass
	version.Libraries = append(version.Libraries, version.Client())

	watcher.Handle(MetadataResolvedEvent{})

	installedLibs, requiredLibs := filterLibraries(version.Libraries)

	assetIndex, err := meta.DownloadAssetIndex(version)
	if err != nil {
		return LaunchEnvironment{}, fmt.Errorf("retrieve asset index: %w", err)
	}

	var downloads []network.DownloadEntry

	if !options.skipLibraries {
		for _, library := range requiredLibs {
			downloads = append(downloads, library.Artifact.DownloadEntry())
		}
	}
	watcher.Handle(LibrariesResolvedEvent{
		Libraries: len(installedLibs) + len(requiredLibs),
	})

	if !options.skipAssets {
		for _, object := range assetIndex.Objects {
			if !object.IsDownloaded() {
				downloads = append(downloads, object.DownloadEntry())
			}
		}
	}
	watcher.Handle(AssetsResolvedEvent{Assets: len(assetIndex.Objects)})

	if len(downloads) > 0 {
		results := network.StartDownloadEntries(downloads)
		i := 0
		for err := range results {
			if err != nil {
				return LaunchEnvironment{}, fmt.Errorf("download files: %w", err)
			}
			watcher.Handle(DownloadingEvent{
				Completed: i,
				Total:     len(downloads),
			})
			i++
		}
	}

	if inst.Loader == LoaderNeoForge || inst.Loader == LoaderForge {
		var post []meta.ForgeProcessor

		if inst.Loader == LoaderForge {
			post, err = meta.Forge.FetchPostProcessors(version.ID, version.LoaderID)
			if err != nil {
				return LaunchEnvironment{}, fmt.Errorf("fetch forge post processors: %w", err)
			}
		} else if inst.Loader == LoaderNeoForge {
			post, err = meta.Neoforge.FetchPostProcessors(version.ID, version.LoaderID)
			if err != nil {
				return LaunchEnvironment{}, fmt.Errorf("fetch neoforge post processors: %w", err)
			}
		}

		for _, processor := range post {
			cmd := exec.Command(launchEnv.JavaPath, processor.JavaArgs...)
			cmd.Dir = inst.Dir()
			cmd.Stderr = os.Stdout
			if err := cmd.Run(); err != nil {
				return LaunchEnvironment{}, fmt.Errorf("run forge post processor: %w", err)
			}
		}
	}

	if runtime.GOOS == "darwin" {
		launchEnv.JavaArgs = append(launchEnv.JavaArgs, "-XstartOnFirstThread")
	}
	if options.Config.MinMemory != 0 {
		launchEnv.JavaArgs = append(launchEnv.JavaArgs, fmt.Sprintf("-Xms%dm", options.Config.MinMemory))
	}
	if options.Config.MaxMemory != 0 {
		launchEnv.JavaArgs = append(launchEnv.JavaArgs, fmt.Sprintf("-Xmx%dm", options.Config.MaxMemory))
	}
	launchEnv.GameArgs = []string{
		"--username", options.Session.Username,
		"--accessToken", options.Session.AccessToken,
		"--userType", "msa",
		"--gameDir", inst.Dir(),
		"--assetsDir", env.AssetsDir,
		"--assetIndex", version.AssetIndex.ID,
		"--version", version.ID,
		"--versionType", version.Type,
		"--width", strconv.Itoa(options.Config.WindowResolution.Width),
		"--height", strconv.Itoa(options.Config.WindowResolution.Height),
	}
	if options.QuickPlayServer != "" {
		launchEnv.GameArgs = append(launchEnv.GameArgs, "--quickPlayMultiplayer", options.QuickPlayServer)
	}
	if options.Session.UUID != "" {
		launchEnv.GameArgs = append(launchEnv.GameArgs, "--uuid", options.Session.UUID)
	}
	if options.Demo {
		launchEnv.GameArgs = append(launchEnv.GameArgs, "--demo")
	}
	if options.DisableChat {
		launchEnv.GameArgs = append(launchEnv.GameArgs, "--disableChat")
	}
	if options.DisableMultiplayer {
		launchEnv.GameArgs = append(launchEnv.GameArgs, "--disableMultiplayer")
	}

	for _, library := range append(installedLibs, requiredLibs...) {
		if library.SkipOnClasspath {
			continue
		}
		launchEnv.Classpath = append(launchEnv.Classpath, library.Artifact.RuntimePath())
	}

	for _, arg := range version.Arguments.Jvm {
		if arg, ok := arg.(string); ok {
			arg = strings.ReplaceAll(arg, "${version_name}", version.ID)
			arg = strings.ReplaceAll(arg, "${library_directory}", env.LibrariesDir)
			arg = strings.ReplaceAll(arg, "${classpath_separator}", string(os.PathListSeparator))
			arg = strings.ReplaceAll(arg, "${classpath_separator}", string(os.PathListSeparator))
			launchEnv.JavaArgs = append(launchEnv.JavaArgs, arg)
		}
	}
	for _, arg := range version.Arguments.Game {
		if arg, ok := arg.(string); ok {
			launchEnv.GameArgs = append(launchEnv.GameArgs, arg)
		}
	}
	return launchEnv, nil
}

// fetchLoaderMeta returns the loader version metadata for the specified loader and version.
func fetchLoaderMeta(loader Loader, gameVersion string, loaderVersion string) (meta.VersionMeta, error) {
	var loaderMeta meta.VersionMeta
	var err error

	if loader == LoaderFabric {
		loaderMeta, err = meta.Fabric.FetchMeta(gameVersion, loaderVersion)
		if err != nil {
			return meta.VersionMeta{}, fmt.Errorf("retrieve fabric metadata: %w", err)
		}
	} else if loader == LoaderQuilt {
		loaderMeta, err = meta.Quilt.FetchMeta(gameVersion, loaderVersion)
		if err != nil {
			return meta.VersionMeta{}, fmt.Errorf("retrieve quilt metadata: %w", err)
		}
	} else if loader == LoaderNeoForge {
		if loaderVersion == "latest" {
			loaderVersion, err = meta.FetchNeoforgeVersion(gameVersion)
			if err != nil {
				return meta.VersionMeta{}, fmt.Errorf("retrieve neoforge version: %w", err)
			}
		}
		loaderMeta, _, err = meta.Neoforge.FetchMeta(loaderVersion)
		if err != nil {
			return meta.VersionMeta{}, fmt.Errorf("retrieve neoforge metadata: %w", err)
		}
	} else if loader == LoaderForge {
		if loaderVersion == "latest" {
			loaderVersion, err = meta.FetchForgeVersion(gameVersion)
			if err != nil {
				return meta.VersionMeta{}, fmt.Errorf("retrieve forge version: %w", err)
			}
		}
		loaderMeta, _, err = meta.Forge.FetchMeta(loaderVersion)
		if err != nil {
			return meta.VersionMeta{}, fmt.Errorf("retrieve forge metadata: %w", err)
		}
	}
	return loaderMeta, nil
}
