package cmd

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/telecter/cmd-launcher/internal/cli/output"
	"github.com/telecter/cmd-launcher/internal/meta"
)

// SearchCmd enables search of game and mod loader versions.
type SearchCmd struct {
	Query   string `arg:"" help:"${search_arg_query}" optional:""`
	Kind    string `help:"${search_arg_kind}" short:"k" enum:"versions,fabric,quilt,forge" default:"versions"`
	Reverse bool   `short:"r" help:"${search_arg_reverse}"`
}

func (c *SearchCmd) Run(ctx *kong.Context) error {
	var rows []table.Row
	var header table.Row

	switch c.Kind {
	case "versions":
		header = table.Row{
			output.Translate("search.table.version"),
			output.Translate("search.table.type"),
			output.Translate("search.table.date"),
		}
		manifest, err := meta.FetchVersionManifest()
		if err != nil {
			return fmt.Errorf("retrieve version manifest: %w", err)
		}
		for _, version := range manifest.Versions {
			if strings.Contains(version.ID, c.Query) {
				rows = append(rows, table.Row{version.ID, version.Type, version.ReleaseTime.Format(time.DateTime)})
			}
		}
	case "fabric", "quilt":
		header = table.Row{
			output.Translate("search.table.version"),
		}
		var versions meta.FabricVersionList

		api := meta.Fabric
		if c.Kind == "quilt" {
			api = meta.Quilt
		}

		versions, err := api.FetchVersions()
		if err != nil {
			return fmt.Errorf("retrieve versions: %w", err)
		}
		for _, version := range versions {
			if strings.Contains(version.Version, c.Query) {
				rows = append(rows, table.Row{version.Version})
			}
		}
	case "forge":
		header = table.Row{
			output.Translate("search.table.version"),
			"Game Version",
			"Type",
		}

		versions, err := meta.FetchForgePromotions()
		if err != nil {
			return fmt.Errorf("retrieve Forge versions: %w", err)
		}
		for _, gameVersion := range versions.Keys() {
			version, _ := versions.Get(gameVersion)
			parts := strings.Split(gameVersion, "-")
			if len(parts) < 2 {
				continue
			}
			if strings.Contains(parts[0], c.Query) {
				rows = append(rows, table.Row{version.(string), parts[0], parts[1]})
			}
		}
	}

	if c.Reverse {
		slices.Reverse(rows)
	}

	output.Success(output.Translate("search.complete"), len(rows))
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(header)
	t.AppendRows(rows)
	t.Render()
	return nil
}
