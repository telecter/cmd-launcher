package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/telecter/cmd-launcher/cmd"
	"github.com/telecter/cmd-launcher/internal"
	"github.com/urfave/cli/v3"
)

var home, _ = os.UserHomeDir()

var app = cli.Command{
	Name:  "cmd-launcher",
	Usage: "A minimal command line Minecraft launcher.",
	Commands: []*cli.Command{
		cmd.Start,
		cmd.Auth,
		cmd.Create,
		cmd.Delete,
		cmd.Search,
	},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "dir",
			Usage: "Root directory to use for launcher",
			Value: filepath.Join(home, ".minecraft"),
		},
		&cli.BoolFlag{
			Name:  "clear-caches",
			Usage: "Clear all caches. Use this flag to see new updates and metadata.",
			Value: false,
			Action: func(ctx context.Context, c *cli.Command, b bool) error {
				if err := os.RemoveAll(internal.CachesDir); err != nil {
					return cli.Exit(fmt.Errorf("failed to clear caches: %w", err), 1)
				}
				return nil
			},
		},
	},
	Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
		internal.RootDir = c.String("dir")
		internal.InstancesDir = filepath.Join(internal.RootDir, "instances")
		internal.LibrariesDir = filepath.Join(internal.RootDir, "libraries")
		internal.CachesDir = filepath.Join(internal.RootDir, "caches")
		internal.AssetsDir = filepath.Join(internal.RootDir, "assets")
		internal.AccountDataCache = filepath.Join(internal.RootDir, "account.json")
		return nil, nil
	},
	ExitErrHandler: func(ctx context.Context, c *cli.Command, err error) {
		log.Fatalln(fmt.Errorf("error: %w", err))
		cli.OsExiter(err.(cli.ExitCoder).ExitCode())
	},
}

func main() {
	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatalln(err)
	}
}
