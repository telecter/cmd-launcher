package launcher

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/telecter/cmd-launcher/internal/meta"
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
				Loader:      meta.LoaderVanilla,
			},
			wantError: false,
		},
		{
			name: "Vanilla Invalid Version",
			options: InstanceOptions{
				GameVersion: "not valid",
				Loader:      meta.LoaderVanilla,
			},
			wantError: true,
		},
		{
			name: "Fabric",
			options: InstanceOptions{
				GameVersion:   "release",
				Loader:        meta.LoaderFabric,
				LoaderVersion: "latest",
			},
			wantError: false,
		},
		{
			name: "Fabric Versioned",
			options: InstanceOptions{
				GameVersion:   "release",
				Loader:        meta.LoaderFabric,
				LoaderVersion: "0.16.14",
			},
			wantError: false,
		},
		{
			name: "Fabric Invalid Version",
			options: InstanceOptions{
				GameVersion:   "release",
				Loader:        meta.LoaderFabric,
				LoaderVersion: "not valid",
			},
			wantError: true,
		},
		{
			name: "Quilt",
			options: InstanceOptions{
				GameVersion:   "release",
				Loader:        meta.LoaderQuilt,
				LoaderVersion: "latest",
			},
			wantError: false,
		},
		{
			name: "Quilt Versioned",
			options: InstanceOptions{
				GameVersion:   "release",
				Loader:        meta.LoaderQuilt,
				LoaderVersion: "0.29.0-beta.7",
			},
			wantError: false,
		},
		{
			name: "Quilt Invalid Version",
			options: InstanceOptions{
				GameVersion:   "release",
				Loader:        meta.LoaderQuilt,
				LoaderVersion: "not valid",
			},
			wantError: true,
		},
		{
			name: "Forge",
			options: InstanceOptions{
				GameVersion:   "release",
				Loader:        meta.LoaderForge,
				LoaderVersion: "latest",
			},
			wantError: false,
		},
		{
			name: "Forge Versioned",
			options: InstanceOptions{
				GameVersion:   "1.21.5",
				Loader:        meta.LoaderForge,
				LoaderVersion: "1.21.5-55.0.22",
			},
			wantError: false,
		},
		{
			name: "Forge Invalid Version",
			options: InstanceOptions{
				GameVersion:   "release",
				Loader:        meta.LoaderForge,
				LoaderVersion: "not valid",
			},
			wantError: true,
		},
		{
			name: "NeoForge",
			options: InstanceOptions{
				GameVersion:   "release",
				Loader:        meta.LoaderNeoForge,
				LoaderVersion: "latest",
			},
			wantError: false,
		},
		{
			name: "NeoForge Versioned",
			options: InstanceOptions{
				GameVersion:   "release",
				Loader:        meta.LoaderNeoForge,
				LoaderVersion: "21.5.75",
			},
			wantError: false,
		},
		{
			name: "NeoForge Invalid Version",
			options: InstanceOptions{
				GameVersion:   "release",
				Loader:        meta.LoaderNeoForge,
				LoaderVersion: "not valid",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.options.Name = uuid.NewString()
			inst, err := CreateInstance(tt.options)
			isError := err != nil
			if tt.wantError != isError {
				t.Errorf("got error: %t; wanted error to be: %t", isError, tt.wantError)
				if isError {
					t.Log(err)
				}
			}
			if _, err := os.Stat(inst.Dir()); err != nil {
				t.Errorf("instance directory should be accessible; got error: %s", err)
			}
		})
	}
}

func TestFetchAllInstances(t *testing.T) {
	env.SetDirs(t.TempDir())
	_, err := CreateInstance(InstanceOptions{
		Name:        uuid.NewString(),
		GameVersion: "release",
		Loader:      meta.LoaderVanilla,
	})
	if err != nil {
		t.Fatalf("unexpected error creating instance for test: %s", err)
	}
	insts, err := FetchAllInstances()
	if err != nil {
		t.Errorf("wanted no error; got: %s", err)
	}
	if len(insts) != 1 {
		t.Errorf("wanted number of instances to be 1; got %d", len(insts))
	}
}

func TestRemoveInstance(t *testing.T) {
	env.SetDirs(t.TempDir())
	inst, err := CreateInstance(InstanceOptions{
		Name:        uuid.NewString(),
		GameVersion: "release",
		Loader:      meta.LoaderVanilla,
	})
	if err != nil {
		t.Fatalf("unexpected error creating instance for test: %s", err)
	}
	if err := RemoveInstance(inst.Name); err != nil {
		t.Errorf("wanted no error; got: %s", err)
	}
	if _, err := os.Stat(inst.Dir()); err == nil {
		t.Error("instance directory should not exist; but does")
	}
}

func TestRenameInstance(t *testing.T) {
	env.SetDirs(t.TempDir())
	inst, err := CreateInstance(InstanceOptions{
		Name:        uuid.NewString(),
		GameVersion: "release",
		Loader:      meta.LoaderVanilla,
	})
	if err != nil {
		t.Fatalf("unexpected error creating instance for test: %s", err)
	}
	name := uuid.NewString()
	if err := inst.Rename(name); err != nil {
		t.Errorf("wanted no error; got: %s", err)
	}
	if _, err := os.Stat(filepath.Join(env.InstancesDir, name)); err != nil {
		t.Error("renamed instance directory does not exist; but should")
	}

}

func testingWatcher(event any) {
	switch e := event.(type) {
	case AssetsResolvedEvent:
		fmt.Printf("Identified %d assets\n", e.Total)
	case LibrariesResolvedEvent:
		fmt.Printf("Identified %d libraries\n", e.Total)

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
				Loader:      meta.LoaderVanilla,
				Config: InstanceConfig{
					Java: "java",
				},
			},
		},
		{
			name: "Vanilla with Mojang JVM",
			options: InstanceOptions{
				GameVersion: "release",
				Loader:      meta.LoaderVanilla,
			},
		},
		{
			name: "Fabric",
			options: InstanceOptions{
				GameVersion:   "release",
				Loader:        meta.LoaderFabric,
				LoaderVersion: "latest",
				Config: InstanceConfig{
					Java: "java",
				},
			},
		},
		{
			name: "Quilt",
			options: InstanceOptions{
				GameVersion:   "release",
				Loader:        meta.LoaderQuilt,
				LoaderVersion: "latest",
				Config: InstanceConfig{
					Java: "java",
				},
			},
		},
		{
			name: "Forge",
			options: InstanceOptions{
				GameVersion:   "release",
				Loader:        meta.LoaderForge,
				LoaderVersion: "latest",
				Config: InstanceConfig{
					Java: "java",
				},
			},
		},
		{
			name: "NeoForge",
			options: InstanceOptions{
				GameVersion:   "release",
				Loader:        meta.LoaderNeoForge,
				LoaderVersion: "latest",
				Config: InstanceConfig{
					Java: "java",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.options.Name = uuid.NewString()
			inst, err := CreateInstance(tt.options)
			if err != nil {
				t.Fatalf("unexpected error creating instance for test: %s", err)
			}

			_, err = Prepare(inst, LaunchOptions{
				Session: auth.Session{
					Username: "testing",
				},
				InstanceConfig: inst.Config,
				skipAssets:     true,
			}, testingWatcher)
			if err != nil {
				t.Errorf("wanted no error; got: %s", err)
			}
		})
	}
}
