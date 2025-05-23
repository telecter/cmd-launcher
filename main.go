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
	},
	Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
		internal.SetDirsFromRoot(c.String("dir"))
		return nil, nil
	},
	ExitErrHandler: func(ctx context.Context, c *cli.Command, err error) {
		log.Fatalln(fmt.Errorf("error: %w", err))
	},
}

func main() {
	app.Run(context.Background(), os.Args)
}
