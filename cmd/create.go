package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/telecter/cmd-launcher/internal/launcher"
	"github.com/urfave/cli/v3"
)

func create(ctx context.Context, c *cli.Command) error {
	if c.Args().Len() < 1 {
		cli.ShowSubcommandHelpAndExit(c, 1)
	}
	instance, err := launcher.CreateInstance(launcher.InstanceOptions{
		GameVersion: c.String("version"),
		Name:        c.Args().First(),
		ModLoader:   c.String("loader"),
	})
	if err != nil {
		return cli.Exit(fmt.Errorf("failed to create instance: %w", err), 1)
	}
	log.Printf("Created instance '%s' with Minecraft %s %s\n", instance.Name, instance.GameVersion, instance.ModLoader)
	return nil
}

var Create = &cli.Command{
	Name:      "create",
	Usage:     "Create a new Minecraft instance",
	ArgsUsage: "<id>",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "loader",
			Usage:   "Mod loader to use",
			Aliases: []string{"l"},
		},
		&cli.StringFlag{
			Name:    "version",
			Usage:   "Game version",
			Aliases: []string{"v"},
			Value:   "release",
		},
	},
	Action: create,
}
