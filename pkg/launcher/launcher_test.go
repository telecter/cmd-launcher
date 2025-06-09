package launcher

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	env "github.com/telecter/cmd-launcher/pkg"
	"github.com/telecter/cmd-launcher/pkg/auth"
)

func TestCreateInstance(t *testing.T) {
	env.SetDirs(t.TempDir())

	tests := []struct {
		name      string
		options   InstanceOptions
		wantError bool
	}{
		{
			name: "Vanilla",
			options: InstanceOptions{
				GameVersion: "release",
				Name:        uuid.NewString(),
				Loader:      LoaderVanilla,
			},
			wantError: false,
		},
		{
			name: "Vanilla Invalid Version",
			options: InstanceOptions{
				GameVersion: "not valid",
				Name:        uuid.NewString(),
				Loader:      LoaderVanilla,
			},
			wantError: true,
		},
		{
			name: "Fabric",
			options: InstanceOptions{
				GameVersion:   "release",
				Name:          uuid.NewString(),
				Loader:        LoaderFabric,
				LoaderVersion: "latest",
			},
			wantError: false,
		},
		{
			name: "Fabric Versioned",
			options: InstanceOptions{
				GameVersion:   "release",
				Name:          uuid.NewString(),
				Loader:        LoaderFabric,
				LoaderVersion: "0.16.14",
			},
			wantError: false,
		},
		{
			name: "Fabric Invalid Version",
			options: InstanceOptions{
				GameVersion:   "release",
				Name:          uuid.NewString(),
				Loader:        LoaderFabric,
				LoaderVersion: "not valid",
			},
			wantError: true,
		},
		{
			name: "Quilt",
			options: InstanceOptions{
				GameVersion:   "release",
				Name:          uuid.NewString(),
				Loader:        LoaderQuilt,
				LoaderVersion: "latest",
			},
			wantError: false,
		},
		{
			name: "Quilt Versioned",
			options: InstanceOptions{
				GameVersion:   "release",
				Name:          uuid.NewString(),
				Loader:        LoaderQuilt,
				LoaderVersion: "0.29.0-beta.7",
			},
			wantError: false,
		},
		{
			name: "Quilt Invalid Version",
			options: InstanceOptions{
				GameVersion:   "release",
				Name:          uuid.NewString(),
				Loader:        LoaderQuilt,
				LoaderVersion: "not valid",
			},
			wantError: true,
		},
		{
			name: "Forge",
			options: InstanceOptions{
				GameVersion: "release",
				Name:        uuid.NewString(),
				Loader:      LoaderForge,
			},
		},
		{
			name: "NeoForge",
			options: InstanceOptions{
				GameVersion: "release",
				Name:        uuid.NewString(),
				Loader:      LoaderNeoForge,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CreateInstance(tt.options)
			isError := err != nil
			if tt.wantError != isError {
				t.Errorf("got error: %t; wanted error to be: %t", isError, tt.wantError)
				if isError {
					t.Log(err)
				}
			}
		})
	}
}

func TestRemoveInstance(t *testing.T) {
	env.SetDirs(t.TempDir())
	inst, err := CreateInstance(InstanceOptions{
		Name:        uuid.NewString(),
		GameVersion: "release",
		Loader:      LoaderVanilla,
	})
	if err != nil {
		t.Fatalf("unexpected error creating instance for test: %s", err)
	}
	if err := RemoveInstance(inst.Name); err != nil {
		t.Errorf("wanted no error; got: %s", err)
	}
}

func TestRenameInstance(t *testing.T) {
	env.SetDirs(t.TempDir())
	inst, err := CreateInstance(InstanceOptions{
		Name:        uuid.NewString(),
		GameVersion: "release",
		Loader:      LoaderVanilla,
	})
	if err != nil {
		t.Fatalf("unexpected error creating instance for test: %s", err)
	}
	if err := inst.Rename(uuid.NewString()); err != nil {
		t.Errorf("wanted no error; got: %s", err)
	}
}

type testingWatcher struct{}

func (watcher testingWatcher) Handle(event any) {
	switch e := event.(type) {
	case AssetsResolvedEvent:
		fmt.Printf("Identified %d assets\n", e.Assets)
	case LibrariesResolvedEvent:
		fmt.Printf("Identified %d libraries\n", e.Libraries)

	case MetadataResolvedEvent:
		fmt.Println("Version metadata retrieved")

	}
}

func TestPrepare(t *testing.T) {
	env.SetDirs(t.TempDir())
	tests := []struct {
		name    string
		options InstanceOptions
	}{
		{
			name: "Vanilla",
			options: InstanceOptions{
				GameVersion: "release",
				Name:        uuid.NewString(),
				Loader:      LoaderVanilla,
			},
		},
		{
			name: "Fabric",
			options: InstanceOptions{
				GameVersion:   "release",
				Name:          uuid.NewString(),
				Loader:        LoaderFabric,
				LoaderVersion: "latest",
			},
		},
		{
			name: "Quilt",
			options: InstanceOptions{
				GameVersion:   "release",
				Name:          uuid.NewString(),
				Loader:        LoaderQuilt,
				LoaderVersion: "latest",
			},
		},
		{
			name: "Forge",
			options: InstanceOptions{
				GameVersion: "release",
				Name:        uuid.NewString(),
				Loader:      LoaderForge,
			},
		},
		{
			name: "NeoForge",
			options: InstanceOptions{
				GameVersion: "release",
				Name:        uuid.NewString(),
				Loader:      LoaderNeoForge,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst, err := CreateInstance(tt.options)
			if err != nil {
				t.Fatalf("unexpected error creating instance for test: %s", err)
			}

			_, err = inst.Prepare(EnvOptions{
				Session: auth.Session{
					Username: "testing",
				},
				Config:     inst.Config,
				skipAssets: true,
			}, testingWatcher{})
			if err != nil {
				t.Errorf("wanted no error; got: %s", err)
			}
		})
	}
}
