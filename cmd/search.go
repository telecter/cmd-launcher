package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/telecter/cmd-launcher/internal/launcher"
	"github.com/telecter/cmd-launcher/internal/network/api"
	"github.com/urfave/cli/v3"
)

func searchInstance(ctx context.Context, c *cli.Command) error {
	instances, err := launcher.GetAllInstances()
	if err != nil {
		return cli.Exit(fmt.Errorf("failed to get all instances"), 1)
	}

	var rows []table.Row
	for i, instance := range instances {
		if strings.Contains(instance.Name, c.Args().First()) {
			rows = append(rows, table.Row{i, instance.Name, instance.GameVersion, instance.ModLoader})
		}
	}
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Name", "Version", "Flags"})
	t.AppendRows(rows)
	t.Render()
	return nil
}
func searchVersions(ctx context.Context, c *cli.Command) error {
	manifest, err := api.GetVersionManifest()
	if err != nil {
		return cli.Exit(fmt.Errorf("failed to search for versions: %w", err), 1)
	}

	var rows []table.Row
	for i, version := range manifest.Versions {
		if strings.Contains(version.ID, c.Args().First()) {
			rows = append(rows, table.Row{i, version.ID, version.Type, version.ReleaseTime.Format(time.UnixDate)})
		}
	}
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Version", "Type", "Release Date"})
	t.AppendRows(rows)
	t.Render()
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
