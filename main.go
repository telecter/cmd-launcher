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
		},
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			err := os.MkdirAll(c.String("dir"), 0755)
			if err != nil {
				return nil, cli.Exit(fmt.Errorf("failed to create launcher directory: %w", err), 1)
			}

			env.RootDir = c.String("dir")
			env.InstancesDir = filepath.Join(env.RootDir, "instances")
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
