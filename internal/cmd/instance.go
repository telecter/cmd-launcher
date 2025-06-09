package cmd

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/telecter/cmd-launcher/pkg/launcher"
)

type Create struct {
	ID            string `arg:"" name:"id" help:"Instance name"`
	Loader        string `name:"loader" help:"Mod loader to use" enum:"fabric,quilt,neoforge,forge,vanilla" default:"vanilla" short:"l"`
	Version       string `name:"version" help:"Game version" default:"release" short:"v"`
	LoaderVersion string `name:"loader-version" help:"Loader version (if Fabric/Quilt)" default:"latest"`
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
	fmt.Printf("Created instance '%s' with Minecraft %s ", inst.Name, inst.GameVersion)
	if inst.Loader != launcher.LoaderVanilla {
		fmt.Printf("(%s %s)\n", inst.Loader, inst.LoaderVersion)
	} else {
		fmt.Print("\n")
	}
	if inst.Config.Java == "" {
		fmt.Println("Java could not be found. This means you will need to manually set its path in the instance configuration.")
	}
	return nil
}

type Delete struct {
	ID  string `arg:"" name:"id" help:"Instance to delete"`
	Yes bool   `name:"yes" short:"y" help:"Assume yes to all questions"`
}

func (c *Delete) Run(ctx *kong.Context) error {
	inst, err := launcher.FetchInstance(c.ID)
	if err != nil {
		return err
	}
	delete := c.Yes
	if !delete {
		var input string
		fmt.Printf("Are you sure you want to delete '%s'?\nIt will be gone forever (a long time!) [y/n] ", inst.Name)
		fmt.Scanln(&input)
		delete = input == "y" || input == "Y"
	}
	if delete {
		if err := launcher.RemoveInstance(c.ID); err != nil {
			return fmt.Errorf("remove instance: %w", err)
		}
		fmt.Printf("Deleted instance '%s'\n", inst.Name)
	} else {
		fmt.Println("Operation aborted")
	}
	return nil
}

type Rename struct {
	ID  string `arg:"" name:"id" help:"Instance to rename"`
	New string `arg:"" name:"new" help:"New name for instance"`
}

func (c *Rename) Run(ctx *kong.Context) error {
	inst, err := launcher.FetchInstance(c.ID)
	if err != nil {
		return err
	}
	if err := inst.Rename(c.New); err != nil {
		return fmt.Errorf("rename instance: %w", err)
	}
	return nil
}

type List struct{}

func (c *List) Run(ctx *kong.Context) error {
	var rows []table.Row
	instances, err := launcher.FetchAllInstances()
	if err != nil {
		return fmt.Errorf("get all instances: %w", err)
	}
	for i, instance := range instances {
		rows = append(rows, table.Row{i, instance.Name, instance.GameVersion, instance.Loader})
	}

	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Name", "Version", "Type"})
	t.AppendRows(rows)
	t.Render()
	return nil
}

type Instance struct {
	Create Create `cmd:"" help:"Create a new instance"`
	Delete Delete `cmd:"" help:"Delete an instance"`
	Rename Rename `cmd:"" help:"Rename an instance"`
	List   List   `cmd:"" help:"List all instances"`
}
