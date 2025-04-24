package launcher

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/telecter/cmd-launcher/internal/auth"
)

type LaunchOptions struct {
	QuickPlayServer string
	LoginData       auth.MinecraftLoginData
	OfflineMode     bool
}

func run(java_path string, args []string) error {
	cmd := exec.Command(java_path, args...)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	go func() {
		io.Copy(os.Stdout, stdout)
	}()
	go func() {
		io.Copy(os.Stderr, stderr)
	}()

	return cmd.Run()
}

func Launch(instanceId string, options LaunchOptions) error {
	instance, err := GetInstance(instanceId)
	if err != nil {
		return err
	}

	if !options.OfflineMode {
		loginData, err := auth.LoginWithMicrosoft()
		if err != nil {
			return fmt.Errorf("error logging in with Microsoft: %w", err)
		}
		options.LoginData = loginData
	}

	meta, err := instance.GetVersionMeta()
	if err != nil {
		return err
	}

	libraries, err := installLibraries(meta.Libraries)
	if err != nil {
		return err
	}

	if instance.ModLoader == "fabric" {
		fabricMeta, err := instance.GetFabricMeta()
		if err != nil {
			return err
		}

		fabricLibraries, err := installLibraries(fabricMeta.Libraries)
		if err != nil {
			return err
		}
		libraries = append(libraries, fabricLibraries...)
	}

	if err = downloadAssets(meta); err != nil {
		return err
	}
	if err := instance.DownloadClient(); err != nil {
		return err
	}

	return instance.Start(libraries, options)
}
