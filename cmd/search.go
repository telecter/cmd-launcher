package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/telecter/cmd-launcher/internal/launcher"
	"github.com/telecter/cmd-launcher/internal/meta"
)

type Search struct {
	Query string `arg:"" name:"query" help:"Search query" optional:""`
	Kind  string `name:"kind" help:"What to search for" short:"k" enum:"instances,versions,fabric,quilt" default:"instances"`
}

func (c *Search) Run(ctx *kong.Context) error {
	var rows []table.Row
	var header table.Row

	switch c.Kind {
	case "instances":
		header = table.Row{"#", "Name", "Version", "Type"}
		instances, err := launcher.GetAllInstances()
		if err != nil {
			return fmt.Errorf("get all instances: %w", err)
		}
		for i, instance := range instances {
			if strings.Contains(instance.Name, c.Query) {
				rows = append(rows, table.Row{i, instance.Name, instance.GameVersion, instance.Loader})
			}
		}
	case "versions":
		header = table.Row{"#", "Version", "Type", "Release Date"}
		manifest, err := meta.GetVersionManifest()
		if err != nil {
			return fmt.Errorf("search for versions: %w", err)
		}
		for i, version := range manifest.Versions {
			if strings.Contains(version.ID, c.Query) {
				rows = append(rows, table.Row{i, version.ID, version.Type, version.ReleaseTime.Format(time.UnixDate)})
			}
		}
	case "fabric", "quilt":
		header = table.Row{"#", "Version"}
		fabricLoader := meta.FabricLoaderStandard
		if c.Kind == "quilt" {
			fabricLoader = meta.FabricLoaderQuilt
		}
		versions, err := meta.GetFabricVersions(fabricLoader)
		if err != nil {
			return fmt.Errorf("search for versions: %w", err)
		}

		for i, version := range versions {
			rows = append(rows, table.Row{i, version.Version})
		}
	}
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(header)
	t.AppendRows(rows)
	t.Render()
	return nil
}
