package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
	"github.com/telecter/cmd-launcher/internal/cli"
	"github.com/telecter/cmd-launcher/pkg/auth"
	"github.com/telecter/cmd-launcher/pkg/launcher"
)

func watcher(verbosity int) launcher.EventWatcher {
	var bar = progressbar.NewOptions(0,
		progressbar.OptionSetDescription(cli.Translate("start.launch.downloading")),
		progressbar.OptionSetWriter(os.Stdout),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionOnCompletion(func() {
			fmt.Print("\n")
		}),
		progressbar.OptionFullWidth())
	return func(event any) {
		switch e := event.(type) {
		case launcher.DownloadingEvent:
			bar.ChangeMax(e.Total)
			bar.Add(1)
		case launcher.AssetsResolvedEvent:
			if verbosity > 0 {
				cli.Info(cli.Translate("start.launch.assets"), e.Total)
			}
		case launcher.LibrariesResolvedEvent:
			if verbosity > 0 {
				cli.Info(cli.Translate("start.launch.libraries"), e.Total)
			}
		case launcher.MetadataResolvedEvent:
			if verbosity > 0 {
				cli.Info(cli.Translate("start.launch.metadata"))
			}
		case launcher.PostProcessingEvent:
			cli.Info(cli.Translate("start.processing"))
		}
	}
}

type StartCmd struct {
	ID string `arg:"" help:"${start_arg_id}"`

	Prepare bool `help:"${start_arg_prepare}"`

	Options struct {
		Username    string `help:"${start_arg_username}" short:"u"`
		Server      string `help:"${start_arg_server}" short:"s" placeholder:"IP" xor:"quickplay"`
		World       string `help:"${start_arg_world}" short:"w" placeholder:"NAME" xor:"quickplay"`
		Demo        bool   `help:"${start_arg_demo}"`
		DisableMP   bool   `help:"${start_arg_disablemp}"`
		DisableChat bool   `help:"${start_arg_disablechat}"`
	} `embed:"" group:"opts"`
	Overrides struct {
		Width     int    `help:"${start_arg_width}" and:"size"`
		Height    int    `help:"${start_arg_height}" and:"size"`
		JVM       string `help:"${start_arg_jvm}" type:"path" placeholder:"PATH"`
		JVMArgs   string `help:"${start_arg_jvmargs}"`
		MinMemory int    `help:"${start_arg_minmemory}" placeholder:"MB" and:"memory"`
		MaxMemory int    `help:"${start_arg_maxmemory}" placeholder:"MB" and:"memory"`
	} `embed:"" group:"overrides"`
}

func (c *StartCmd) Run(ctx *kong.Context, verbosity int) error {
	inst, err := launcher.FetchInstance(c.ID)
	if err != nil {
		return err
	}

	config := inst.Config
	override := launcher.InstanceConfig{
		WindowResolution: struct {
			Width  int "toml:\"width\" json:\"width\""
			Height int "toml:\"height\" json:\"height\""
		}{
			Width:  c.Overrides.Width,
			Height: c.Overrides.Height,
		},
		Java:      c.Overrides.JVM,
		JavaArgs:  c.Overrides.JVMArgs,
		MinMemory: c.Overrides.MinMemory,
		MaxMemory: c.Overrides.MaxMemory,
	}
	if override.WindowResolution.Width != 0 && override.WindowResolution.Height != 0 {
		config.WindowResolution = override.WindowResolution
	}
	if override.Java != "" {
		config.Java = override.Java
	}
	if override.JavaArgs != "" {
		config.JavaArgs = override.JavaArgs
	}
	if override.MinMemory != 0 && override.MaxMemory != 0 {
		config.MinMemory = override.MinMemory
		config.MaxMemory = override.MaxMemory
	}

	session := auth.Session{
		Username: c.Options.Username,
	}
	if c.Options.Username == "" {
		session, err = auth.Authenticate()
		if err != nil {
			return fmt.Errorf("authenticate session: %w", err)
		}
	}

	launchEnv, err := launcher.Prepare(
		inst,
		launcher.LaunchOptions{
			Session: session,

			InstanceConfig:     config,
			QuickPlayServer:    c.Options.Server,
			QuickPlayWorld:     c.Options.World,
			Demo:               c.Options.Demo,
			DisableMultiplayer: c.Options.DisableMP,
			DisableChat:        c.Options.DisableChat,
		},
		watcher(verbosity))

	if err != nil {
		return err
	}

	if c.Prepare {
		cli.Success(cli.Translate("start.prepared"))
		return nil
	}

	if verbosity > 1 {
		cli.Debug(cli.Translate("start.launch.jvmargs"), launchEnv.JavaArgs)

		var gameArgs []string
		var hideNext bool
		for _, arg := range launchEnv.GameArgs {
			if hideNext {
				gameArgs = append(gameArgs, "***")
			} else {
				gameArgs = append(gameArgs, arg)
			}
			if arg == "--accessToken" || arg == "--uuid" {
				hideNext = true
			} else {
				hideNext = false
			}
		}
		cli.Debug(cli.Translate("start.launch.gameargs"), gameArgs)
		cli.Debug(cli.Translate("start.launch.info"), launchEnv.MainClass, launchEnv.GameDir)
	}
	cli.Success(cli.Translate("start.launch"), color.New(color.Bold).Sprint(session.Username))

	return launcher.Launch(launchEnv, launcher.ConsoleRunner)
}
