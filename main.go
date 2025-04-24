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
				Usage: "Clears all caches. Use this flag to see new updates and metadata.",
				Value: false,
			},
		},
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			env.RootDir = c.String("dir")
			env.InstancesDir = filepath.Join(env.RootDir, "instances")
			env.CachesDir = filepath.Join(env.RootDir, "caches")
			env.VersionCachesDir = filepath.Join(env.CachesDir, "minecraft")
			env.FabricCachesDir = filepath.Join(env.CachesDir, "fabric")
			env.AssetsDir = filepath.Join(env.RootDir, "assets")

			if c.Bool("clear-caches") {
				if err := os.RemoveAll(env.CachesDir); err != nil {
					return nil, cli.Exit(fmt.Errorf("failed to clear caches: %w", err), 1)
				}
				log.Println("Cleared all caches")
			}

			if err := os.MkdirAll(c.String("dir"), 0755); err != nil {
				return nil, cli.Exit(fmt.Errorf("failed to create launcher directory: %w", err), 1)
			}
			if err := os.MkdirAll(env.VersionCachesDir, 0755); err != nil {
				return nil, cli.Exit(fmt.Errorf("failed to create Minecraft cache directory: %w", err), 1)
			}
			if err := os.MkdirAll(env.FabricCachesDir, 0755); err != nil {
				return nil, cli.Exit(fmt.Errorf("failed to create Fabric cache directory: %w", err), 1)
			}
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
