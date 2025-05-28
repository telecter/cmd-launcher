package cmd

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/telecter/cmd-launcher/internal/launcher"
)

type Create struct {
	ID      string `arg:"" name:"id" help:"Instance name"`
	Loader  string `name:"loader" help:"Mod loader to use" enum:"fabric,quilt,vanilla" default:"vanilla" short:"l"`
	Version string `name:"version" help:"Game version" default:"release" short:"v"`
}

func (c *Create) Run(ctx *kong.Context) error {
	var loader launcher.Loader
	switch c.Loader {
	case launcher.LoaderFabric.String():
		loader = launcher.LoaderFabric
	case launcher.LoaderQuilt.String():
		loader = launcher.LoaderQuilt
	case launcher.LoaderVanilla.String():
		loader = launcher.LoaderVanilla
	}

	inst, err := launcher.CreateInstance(launcher.InstanceOptions{
		GameVersion: c.Version,
		Name:        c.ID,
		Loader:      loader,
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
	return nil
}
