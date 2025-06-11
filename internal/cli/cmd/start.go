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

type watcher struct {
	progressbar *progressbar.ProgressBar
	verbosity   int
}

func (watcher watcher) Handle(event any) {
	switch e := event.(type) {
	case launcher.DownloadingEvent:
		watcher.progressbar.ChangeMax(e.Total)
		watcher.progressbar.Add(1)
	case launcher.AssetsResolvedEvent:
		if watcher.verbosity > 0 {
			cli.Info(cli.Translate("start.assets", e.Assets))
		}
	case launcher.LibrariesResolvedEvent:
		if watcher.verbosity > 0 {
			cli.Info(cli.Translate("start.libraries", e.Libraries))
		}
	case launcher.MetadataResolvedEvent:
		if watcher.verbosity > 0 {
			cli.Info(cli.Translate("start.metadata"))
		}
	}
}

type Start struct {
	ID string `arg:"" name:"id" help:"${cmd_start_id}"`

	Username    string `help:"${cmd_start_username}" short:"u" group:"opts"`
	Server      string `help:"${cmd_start_server}" short:"s" group:"opts"`
	Demo        bool   `help:"${cmd_start_demo}" group:"opts"`
	DisableMP   bool   `help:"${cmd_start_disablemp}" group:"opts"`
	DisableChat bool   `help:"${cmd_start_disablechat}" group:"opts"`
	Width       int    `help:"${cmd_start_width}" group:"overrides"`
	Height      int    `help:"${cmd_start_height}" group:"overrides" `
	JVM         string `help:"${cmd_start_jvm}" group:"overrides" type:"path" placeholder:"PATH"`
	MinMemory   int    `help:"${cmd_start_minmemory}" group:"overrides" placeholder:"MB"`
	MaxMemory   int    `help:"${cmd_start_maxmemory}" group:"overrides" placeholder:"MB"`
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
			Width:  c.Width,
			Height: c.Height,
		},
		Java:      c.JVM,
		MinMemory: c.MinMemory,
		MaxMemory: c.MaxMemory,
	}
	if override.WindowResolution.Width != 0 && override.WindowResolution.Height != 0 {
		config.WindowResolution = override.WindowResolution
	}
	if override.Java != "" {
		config.Java = override.Java
	}
	if override.MinMemory != 0 && override.MaxMemory != 0 {
		config.MinMemory = override.MinMemory
		config.MaxMemory = override.MaxMemory
	}

	session := auth.Session{
		Username: c.Username,
	}
	if c.Username == "" {
		session, err = auth.Authenticate()
		if err != nil {
			return fmt.Errorf("authenticate session: %w", err)
		}
	}

	options := launcher.EnvOptions{
		Session:            session,
		Config:             config,
		QuickPlayServer:    c.Server,
		Demo:               c.Demo,
		DisableMultiplayer: c.DisableMP,
		DisableChat:        c.DisableChat,
	}

	launchEnv, err := launcher.Prepare(inst, options, watcher{
		progressbar: progressbar.NewOptions(0,
			progressbar.OptionSetDescription(cli.Translate("cmd.start.downloading")),
			progressbar.OptionSetWriter(os.Stdout),
			progressbar.OptionThrottle(65*time.Millisecond),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionOnCompletion(func() {
				fmt.Print("\n")
			}),
			progressbar.OptionFullWidth(),
		),
		verbosity: verbosity,
	})

	if err != nil {
		return err
	}

	if verbosity > 1 {
		cli.Debug(cli.Translate("start.debug.jvmargs", launchEnv.JavaArgs))

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
		cli.Debug(cli.Translate("start.debug.gameargs", gameArgs))
		cli.Debug(cli.Translate("start.debug.info", launchEnv.MainClass, launchEnv.GameDir))
	}

	cli.Success(cli.Translate("start.launching", session.Username))

	return launcher.Launch(launchEnv, launcher.ConsoleRunner{})
}
