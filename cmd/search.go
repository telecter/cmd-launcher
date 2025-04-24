package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/telecter/cmd-launcher/internal/launcher"
	"github.com/telecter/cmd-launcher/internal/network/api"
	"github.com/urfave/cli/v3"
)

func searchInstance(ctx context.Context, c *cli.Command) error {
	query := c.Args().First()

	instances, err := launcher.GetAllInstances()
	if err != nil {
		return cli.Exit(fmt.Errorf("failed to get all instances"), 1)
	}
	for _, instance := range instances {
		if strings.Contains(instance.Name, query) {
			fmt.Printf("%s (%s) %s\n", instance.Name, instance.GameVersion, instance.ModLoader)
		}
	}
	return nil
}
func searchVersions(ctx context.Context, c *cli.Command) error {
	query := c.Args().First()

	manifest, err := api.GetVersionManifest()
	if err != nil {
		return cli.Exit(fmt.Errorf("failed to search for versions: %w", err), 1)
	}

	for _, version := range manifest.Versions {
		if strings.Contains(version.ID, query) {
			fmt.Printf("%s (%s) released at %s\n", version.ID, version.Type, version.ReleaseTime.Format(time.UnixDate))
		}
	}

	return nil
}

var Search = &cli.Command{
	Name:  "search",
	Usage: "Search versions and instances",
	Commands: []*cli.Command{
		{
			Name:      "instances",
			Usage:     "Search for an instance",
			ArgsUsage: "[instance]",
			Action:    searchInstance,
		},
		{
			Name:      "versions",
			Usage:     "Search for Minecraft versions",
			ArgsUsage: "[version]",
			Action:    searchVersions,
		},
	},
}
