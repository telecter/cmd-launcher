package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/telecter/cmd-launcher/cmd"
	"github.com/urfave/cli/v2"
)

var home, _ = os.UserHomeDir()

func main() {
	app := cli.App{
		Name:  "cmd-launcher",
		Usage: "A minimal command line Minecraft launcher.",
		Commands: []*cli.Command{
			cmd.Start,
			cmd.Mod,
			cmd.Auth,
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "dir",
				Usage: "Root directory to use for launcher",
				Value: filepath.Join(home, ".minecraft"),
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
