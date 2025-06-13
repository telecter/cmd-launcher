package cmd

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/telecter/cmd-launcher/internal/cli"
	"github.com/telecter/cmd-launcher/pkg/launcher"
)

type Create struct {
	ID            string `arg:"" help:"${cmd_create_id}"`
	Loader        string `help:"${cmd_create_loader}" enum:"fabric,quilt,neoforge,forge,vanilla" default:"vanilla" short:"l"`
	Version       string `help:"${cmd_create_version}" default:"release" short:"v"`
	LoaderVersion string `help:"${cmd_create_loaderversion}" default:"latest"`
}

func (c *Create) Run(ctx *kong.Context) error {
	var loader launcher.Loader
	switch c.Loader {
	case "fabric":
		loader = launcher.LoaderFabric
	case "quilt":
		loader = launcher.LoaderQuilt
	case "vanilla":
		loader = launcher.LoaderVanilla
	case "neoforge":
		loader = launcher.LoaderNeoForge
	case "forge":
		loader = launcher.LoaderForge
	}

	inst, err := launcher.CreateInstance(launcher.InstanceOptions{
		GameVersion:   c.Version,
		Name:          c.ID,
		Loader:        loader,
		LoaderVersion: c.LoaderVersion,
	})
	if err != nil {
		return fmt.Errorf("create instance: %w", err)
	}

	l := inst.LoaderVersion
	if l != "" {
		l = " " + l
	}
	cli.Success(cli.Translate("cmd.create.complete"), color.New(color.Bold).Sprint(inst.Name), inst.GameVersion, inst.Loader, l)

	if inst.Config.Java == "" {
		cli.Warning(cli.Translate("cmd.create.nojvm"))
	}
	return nil
}

type Delete struct {
	ID  string `arg:"" name:"id" help:"${cmd_delete_id}"`
	Yes bool   `name:"yes" short:"y" help:"${cmd_delete_yes}"`
}

func (c *Delete) Run(ctx *kong.Context) error {
	inst, err := launcher.FetchInstance(c.ID)
	if err != nil {
		return err
	}
	delete := c.Yes
	if !delete {
		var input string

		cli.Warning(cli.Translate("cmd.delete.confirm"))
		fmt.Printf(cli.Translate("cmd.delete.warning"), color.New(color.Bold).Sprint(inst.Name))
		fmt.Scanln(&input)
		delete = input == "y" || input == "Y"
	}
	if delete {
		if err := launcher.RemoveInstance(c.ID); err != nil {
			return fmt.Errorf("remove instance: %w", err)
		}
		cli.Success(cli.Translate("cmd.delete.complete"), color.New(color.Bold).Sprint(inst.Name))
	} else {
		cli.Info(cli.Translate("cmd.delete.abort"))
	}
	return nil
}

type Rename struct {
	ID  string `arg:"" help:"${cmd_rename_id}"`
	New string `arg:"" help:"${cmd_rename_new}"`
}

func (c *Rename) Run(ctx *kong.Context) error {
	inst, err := launcher.FetchInstance(c.ID)
	if err != nil {
		return err
	}
	if err := inst.Rename(c.New); err != nil {
		return fmt.Errorf("rename instance: %w", err)
	}
	cli.Success(cli.Translate("cmd.rename.complete"))
	return nil
}

type List struct{}

func (c *List) Run(ctx *kong.Context) error {
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
		cli.Translate("search.table.name"),
		cli.Translate("search.table.version"),
		cli.Translate("search.table.type"),
	})
	t.AppendRows(rows)
	t.Render()
	return nil
}

type Instance struct {
	Create Create `cmd:"" help:"${cmd_create}"`
	Delete Delete `cmd:"" help:"${cmd_delete}"`
	Rename Rename `cmd:"" help:"${cmd_rename}"`
	List   List   `cmd:"" help:"${cmd_list}"`
}
