package cmd

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/telecter/cmd-launcher/internal/cli"
	"github.com/telecter/cmd-launcher/internal/meta"
)

type Search struct {
	Query   string `arg:"" help:"${cmd_search_query}" optional:""`
	Kind    string `help:"${cmd_search_kind}" short:"k" enum:"versions,fabric,quilt" default:"versions"`
	Reverse bool   `short:"r" help:"${cmd_search_reverse}"`
}

func (c *Search) Run(ctx *kong.Context) error {
	var rows []table.Row
	var header table.Row

	switch c.Kind {
	case "versions":
		header = table.Row{
			"#",
			cli.Translate("search.table.version"),
			cli.Translate("search.table.type"),
			cli.Translate("search.table.date"),
		}
		manifest, err := meta.FetchVersionManifest()
		if err != nil {
			return fmt.Errorf("retrieve version manifest: %w", err)
		}
		for i, version := range manifest.Versions {
			if strings.Contains(version.ID, c.Query) {
				rows = append(rows, table.Row{i, version.ID, version.Type, version.ReleaseTime.Format(time.UnixDate)})
			}
		}
	case "fabric", "quilt":
		header = table.Row{
			"#",
			cli.Translate("search.table.version"),
		}
		var versions meta.FabricVersionList

		if c.Kind == "fabric" {
			var err error
			versions, err = meta.Fabric.FetchVersions()
			if err != nil {
				return fmt.Errorf("retrieve fabric versions: %w", err)
			}
		} else if c.Kind == "quilt" {
			var err error
			versions, err = meta.Quilt.FetchVersions()
			if err != nil {
				return fmt.Errorf("retrieve quilt versions: %w", err)
			}
		}

		for i, version := range versions {
			if strings.Contains(version.Version, c.Query) {
				rows = append(rows, table.Row{i, version.Version})
			}
		}
	}

	if c.Reverse {
		slices.Reverse(rows)
	}

	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(header)
	t.AppendRows(rows)
	t.Render()
	return nil
}
