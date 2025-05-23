package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/telecter/cmd-launcher/internal/launcher"
	"github.com/urfave/cli/v3"
)

var Create = &cli.Command{
	Name:  "create",
	Usage: "Create a new Minecraft instance",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "loader",
			Usage:   "Mod loader to use",
			Aliases: []string{"l"},
			Value:   "vanilla",
		},
		&cli.StringFlag{
			Name:    "version",
			Usage:   "Game version",
			Aliases: []string{"v"},
			Value:   "release",
		},
	},
	Arguments: []cli.Argument{
		&cli.StringArg{
			Name: "id",
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		if c.StringArg("id") == "" {
			cli.ShowSubcommandHelpAndExit(c, 1)
		}
		inst, err := launcher.CreateInstance(launcher.InstanceOptions{
			GameVersion: c.String("version"),
			Name:        c.StringArg("id"),
			Loader:      c.String("loader"),
		})
		if err != nil {
			return fmt.Errorf("failed to create instance: %w", err)
		}
		log.Printf("Created instance '%s' with Minecraft %s %s\n", inst.Name, inst.GameVersion, inst.Loader)
		return nil
	},
}
