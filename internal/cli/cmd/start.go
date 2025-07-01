package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/schollz/progressbar/v3"
	"github.com/telecter/cmd-launcher/internal/cli"
	"github.com/telecter/cmd-launcher/pkg/auth"
	"github.com/telecter/cmd-launcher/pkg/launcher"
)

func watcher(verbosity int) launcher.EventWatcher {
	var bar = progressbar.NewOptions(0,
		progressbar.OptionSetDescription(cli.Translate("cmd.start.downloading")),
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
				cli.Info(cli.Translate("start.assets"), e.Total)
			}
		case launcher.LibrariesResolvedEvent:
			if verbosity > 0 {
				cli.Info(cli.Translate("start.libraries"), e.Total)
			}
		case launcher.MetadataResolvedEvent:
			if verbosity > 0 {
				cli.Info(cli.Translate("start.metadata"))
			}
		}
	}
}

type Start struct {
	ID string `arg:"" name:"id" help:"${cmd_start_id}"`

	Options struct {
		Username    string `help:"${cmd_start_username}" short:"u"`
		Server      string `help:"${cmd_start_server}" short:"s"`
		Demo        bool   `help:"${cmd_start_demo}"`
		DisableMP   bool   `help:"${cmd_start_disablemp}"`
		DisableChat bool   `help:"${cmd_start_disablechat}"`
	} `embed:"" group:"opts"`
	Overrides struct {
		Width     int    `help:"${cmd_start_width}" and:"size"`
		Height    int    `help:"${cmd_start_height}" and:"size"`
		JVM       string `help:"${cmd_start_jvm}" type:"path" placeholder:"PATH"`
		JVMArgs   string `help:"${cmd_start_jvmargs}"`
		MinMemory int    `help:"${cmd_start_minmemory}" placeholder:"MB" and:"memory"`
		MaxMemory int    `help:"${cmd_start_maxmemory}" placeholder:"MB" and:"memory"`
	} `embed:"" group:"overrides"`
}

func (c *Start) Run(ctx *kong.Context, verbosity int) error {
	inst, err := launcher.FetchInstance(c.ID)
	if err != nil {
		return err
	}

	config := inst.Config
	override := launcher.InstanceConfig{
		WindowResolution: struct {
			Width  int "json:\"width\""
			Height int "json:\"height\""
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
			Demo:               c.Options.Demo,
			DisableMultiplayer: c.Options.DisableMP,
			DisableChat:        c.Options.DisableChat,
		},
		watcher(verbosity))

	if err != nil {
		return err
	}

	if verbosity > 1 {
		cli.Debug(cli.Translate("start.debug.jvmargs"), launchEnv.JavaArgs)

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
		cli.Debug(cli.Translate("start.debug.gameargs"), gameArgs)
		cli.Debug(cli.Translate("start.debug.info"), launchEnv.MainClass, launchEnv.GameDir)
	}
	cli.Success(cli.Translate("start.launching"), session.Username)

	return launcher.Launch(launchEnv, launcher.ConsoleRunner)
}
