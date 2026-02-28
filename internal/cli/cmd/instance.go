package cmd

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/telecter/cmd-launcher/internal/cli/output"
	"github.com/telecter/cmd-launcher/internal/meta"
	"github.com/telecter/cmd-launcher/pkg/launcher"
)

// CreateCmd creates a new instance with specified parameters.
type CreateCmd struct {
	ID            string `arg:"" help:"${create_arg_id}"`
	Loader        string `help:"${create_arg_loader}" enum:"fabric,quilt,neoforge,forge,vanilla" default:"vanilla" short:"l"`
	Version       string `help:"${create_arg_version}" default:"release" short:"v"`
	LoaderVersion string `help:"${create_arg_loaderversion}" default:"latest"`
}

func (c *CreateCmd) Run(ctx *kong.Context) error {
	var loader meta.Loader
	switch c.Loader {
	case "fabric":
		loader = meta.LoaderFabric
	case "quilt":
		loader = meta.LoaderQuilt
	case "vanilla":
		loader = meta.LoaderVanilla
	case "neoforge":
		loader = meta.LoaderNeoForge
	case "forge":
		loader = meta.LoaderForge
	}
	inst, err := launcher.CreateInstance(launcher.InstanceOptions{
		GameVersion:   c.Version,
		Name:          c.ID,
		Loader:        loader,
		LoaderVersion: c.LoaderVersion,
		Config:        defaultInstanceConfig,
	})
	if err != nil {
		return fmt.Errorf("create instance: %w", err)
	}

	l := inst.LoaderVersion
	if l != "" {
		l = " " + l
	}
	output.Success(output.Translate("create.complete"), color.New(color.Bold).Sprint(inst.Name), inst.GameVersion, inst.Loader, l)
	output.Tip(output.Translate("tip.configure"))
	return nil
}

// DeleteCmd removes the specified instance.
type DeleteCmd struct {
	ID  string `arg:"" name:"id" help:"${delete_arg_id}"`
	Yes bool   `name:"yes" short:"y" help:"${delete_arg_yes}"`
}

func (c *DeleteCmd) Run(ctx *kong.Context) error {
	inst, err := launcher.FetchInstance(c.ID)
	if err != nil {
		return err
	}
	delete := c.Yes
	if !delete {
		var input string

		output.Warning(output.Translate("delete.confirm"))
		fmt.Printf(output.Translate("delete.warning"), color.New(color.Bold).Sprint(inst.Name))
		fmt.Scanln(&input)
		delete = input == "y" || input == "Y"
	}
	if delete {
		if err := launcher.RemoveInstance(c.ID); err != nil {
			return fmt.Errorf("remove instance: %w", err)
		}
		output.Success(output.Translate("delete.complete"), color.New(color.Bold).Sprint(inst.Name))
	} else {
		output.Info(output.Translate("delete.abort"))
	}
	return nil
}

// RenameCmd renames the specified instance.
type RenameCmd struct {
	ID  string `arg:"" help:"${rename_arg_id}"`
	New string `arg:"" help:"${rename_arg_new}"`
}

func (c *RenameCmd) Run(ctx *kong.Context) error {
	inst, err := launcher.FetchInstance(c.ID)
	if err != nil {
		return err
	}
	if err := inst.Rename(c.New); err != nil {
		return fmt.Errorf("rename instance: %w", err)
	}
	output.Success(output.Translate("rename.complete"))
	return nil
}

// ListCmd lists all installed instances.
type ListCmd struct{}

func (c *ListCmd) Run(ctx *kong.Context) error {
	var rows []table.Row
	instances, err := launcher.FetchAllInstances()
	if err != nil {
		return fmt.Errorf("fetch all instances: %w", err)
	}
	for i, inst := range instances {
		rows = append(rows, table.Row{i, inst.Name, inst.GameVersion, inst.Loader})
	}

	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{
		"#",
		output.Translate("search.table.name"),
		output.Translate("search.table.version"),
		output.Translate("search.table.type"),
	})
	t.AppendRows(rows)
	t.Render()
	return nil
}

// InstanceCmd enables management of Minecraft instances.
type InstanceCmd struct {
	Create CreateCmd `cmd:"" help:"${create}"`
	Delete DeleteCmd `cmd:"" help:"${delete}"`
	Rename RenameCmd `cmd:"" help:"${rename}"`
	List   ListCmd   `cmd:"" help:"${list}"`
}

var defaultInstanceConfig = launcher.InstanceConfig{
	WindowResolution: struct {
		Width  int "toml:\"width\" json:\"width\""
		Height int "toml:\"height\" json:\"height\""
	}{
		Width:  1708,
		Height: 960,
	},
	MinMemory: 512,
	MaxMemory: 4096,
}
