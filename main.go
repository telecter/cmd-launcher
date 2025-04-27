package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/telecter/cmd-launcher/cmd"
	"github.com/telecter/cmd-launcher/internal/env"
	"github.com/urfave/cli/v3"
)

var home, _ = os.UserHomeDir()

func main() {
	app := cli.Command{
		Name:  "cmd-launcher",
		Usage: "A minimal command line Minecraft launcher.",
		Commands: []*cli.Command{
			cmd.StartCommand,
			cmd.AuthCommand,
			cmd.CreateCommand,
			cmd.DeleteCommand,
			cmd.SearchCommand,
			cmd.ModsCommand,
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "dir",
				Usage: "Root directory to use for launcher",
				Value: filepath.Join(home, ".minecraft"),
			},
			&cli.BoolFlag{
				Name:  "clear-caches",
				Usage: "Clears all caches. Use this flag to see new updates and metadata.",
				Value: false,
				Action: func(ctx context.Context, c *cli.Command, b bool) error {
					if err := os.RemoveAll(env.CachesDir); err != nil {
						return cli.Exit(fmt.Errorf("failed to clear caches: %w", err), 1)
					}
					return nil
				},
			},
		},
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			env.RootDir = c.String("dir")
			env.InstancesDir = filepath.Join(env.RootDir, "instances")
			env.LibrariesDir = filepath.Join(env.RootDir, "libraries")
			env.CachesDir = filepath.Join(env.RootDir, "caches")
			env.AssetsDir = filepath.Join(env.RootDir, "assets")
			env.AccountDataCache = filepath.Join(env.RootDir, "account.json")
			if err := os.MkdirAll(env.InstancesDir, 0755); err != nil {
				return nil, cli.Exit(fmt.Errorf("failed to create instances directory: %w", err), 1)
			}
			return nil, nil
		},
		ExitErrHandler: func(ctx context.Context, c *cli.Command, err error) {
			log.Fatalln(err)
			cli.OsExiter(err.(cli.ExitCoder).ExitCode())
		},
	}
	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatalln(err)
	}
}
