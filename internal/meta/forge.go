package meta

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/iancoleman/orderedmap"
	"github.com/telecter/cmd-launcher/internal/network"
	env "github.com/telecter/cmd-launcher/pkg"
)

// A ForgeInstallProfile contains install libraries and processors used to initialize Forge.
type ForgeInstallProfile struct {
	Spec        int    `json:"spec"`
	Profile     string `json:"profile"`
	Version     string `json:"version"`
	Minecraft   string `json:"minecraft"`
	JSON        string `json:"json"`
	Logo        string `json:"logo"`
	Welcome     string `json:"welcome"`
	MirrorList  string `json:"mirrorList"`
	HideExtract bool   `json:"hideExtract"`
	Data        map[string]struct {
		Client string `json:"client"`
		Server string `json:"server"`
	} `json:"data"`
	Processors []processor `json:"processors"`
	Libraries  []Library   `json:"libraries"`
}
type processor struct {
	Sides     []string           `json:"sides,omitempty"`
	Jar       LibrarySpecifier   `json:"jar"`
	Classpath []LibrarySpecifier `json:"classpath"`
	Args      []string           `json:"args"`
}

// A ForgeProcessor contains the finished JVM arguments to start a Forge post processor.
type ForgeProcessor struct {
	JavaArgs []string
}

// FetchNeoforgeVersion retrieves the best NeoForge loader version for the specified game version.
func FetchNeoforgeVersion(gameVersion string) (string, error) {
	parts := strings.Split(gameVersion, ".")
	var url string
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid version format")
	}

	if gameVersion == "1.20.1" {
		url = "https://maven.neoforged.net/api/maven/latest/version/releases/net/neoforged/forge?filter=1.20.1-"
	} else {
		end := strings.Join(parts[1:], ".")
		url = fmt.Sprintf("https://maven.neoforged.net/api/maven/latest/version/releases/net/neoforged/neoforge?filter=%s", end)
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if err := network.CheckResponse(resp); err != nil {
		var statusErr *network.HTTPStatusError
		if errors.As(err, &statusErr) && statusErr.StatusCode == 404 {
			return "", fmt.Errorf("no version found for specified game version")
		}
		return "", err
	}

	body, _ := io.ReadAll(resp.Body)
	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}
	version, ok := data["version"].(string)
	if !ok {
		return "", fmt.Errorf("invalid data")
	}

	return version, nil
}

func FetchForgePromotions() (*orderedmap.OrderedMap, error) {
	resp, err := http.Get("https://files.minecraftforge.net/net/minecraftforge/forge/promotions_slim.json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := network.CheckResponse(resp); err != nil {
		return nil, err
	}
	body, _ := io.ReadAll(resp.Body)
	var data struct {
		Promos orderedmap.OrderedMap `json:"promos"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("read promoted versions: %w", err)
	}
	return &data.Promos, nil
}

// FetchForgeVersion retrieves the best Forge loader version for the specified game version.
func FetchForgeVersion(gameVersion string) (string, error) {
	type response struct {
		Promos map[string]string `json:"promos"`
	}
	resp, err := http.Get("https://files.minecraftforge.net/net/minecraftforge/forge/promotions_slim.json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if err := network.CheckResponse(resp); err != nil {
		return "", err
	}
	body, _ := io.ReadAll(resp.Body)
	var data response
	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("read promoted versions: %w", err)
	}

	version, ok := data.Promos[gameVersion+"-latest"]
	if !ok {
		return "", fmt.Errorf("no version found for specified game version")
	}
	return gameVersion + "-" + version, nil

}

type forge struct {
	url func(version string) string
}

var Forge = forge{
	url: func(version string) string {
		return fmt.Sprintf("https://maven.minecraftforge.net/net/minecraftforge/forge/%s/forge-%s-installer.jar", version, version)
	},
}
var Neoforge = forge{
	url: func(version string) string {
		return fmt.Sprintf("https://maven.neoforged.net/releases/net/neoforged/neoforge/%s/neoforge-%s-installer.jar", version, version)
	},
}

// FetchInstaller fetchs the Forge installer ZIP file and returns its contents.
func (forge forge) FetchInstaller(version string) (map[string]*zip.File, error) {
	url := forge.url(version)
	path := filepath.Join(env.CachesDir, "forge", path.Base(url))

	if _, err := os.Stat(path); err != nil {
		err := network.DownloadFile(network.DownloadEntry{
			URL:  url,
			Path: path,
		})
		if err != nil {
			var statusErr *network.HTTPStatusError
			if errors.As(err, &statusErr) {
				if statusErr.StatusCode == 404 {
					return nil, fmt.Errorf("invalid version")
				}
			}
			return nil, err
		}
	}

	data, err := os.ReadFile(path)

	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	files := make(map[string]*zip.File)
	for _, file := range r.File {
		files[file.Name] = file
	}
	return files, nil
}

// FetchMeta retrieves the Forge version.json (version meta) and install_profile.json from the installer ZIP.
func (forge forge) FetchMeta(version string) (VersionMeta, ForgeInstallProfile, error) {
	files, err := forge.FetchInstaller(version)

	if err != nil {
		return VersionMeta{}, ForgeInstallProfile{}, fmt.Errorf("fetch installer: %w", err)
	}
	file, ok := files["version.json"]
	if !ok {
		return VersionMeta{}, ForgeInstallProfile{}, fmt.Errorf("version metadata not present in installer")
	}
	rc, err := file.Open()
	if err != nil {
		return VersionMeta{}, ForgeInstallProfile{}, fmt.Errorf("open version metadata: %w", err)
	}
	defer rc.Close()
	c, _ := io.ReadAll(rc)
	var versionMeta VersionMeta
	if err := json.Unmarshal(c, &versionMeta); err != nil {
		return VersionMeta{}, ForgeInstallProfile{}, fmt.Errorf("parse version metadata: %w", err)
	}

	file, ok = files["install_profile.json"]
	if !ok {
		return VersionMeta{}, ForgeInstallProfile{}, fmt.Errorf("install profile not present in installer")
	}
	rc, err = file.Open()
	if err != nil {
		return VersionMeta{}, ForgeInstallProfile{}, fmt.Errorf("open install profile: %w", err)
	}
	defer rc.Close()
	c, _ = io.ReadAll(rc)
	var profile ForgeInstallProfile
	if err := json.Unmarshal(c, &profile); err != nil {
		return VersionMeta{}, ForgeInstallProfile{}, fmt.Errorf("parse install profile: %w", err)
	}
	for _, library := range profile.Libraries {
		library.SkipOnClasspath = true
		versionMeta.Libraries = append(versionMeta.Libraries, library)
	}
	if variable, ok := profile.Data["MC_OFF"]; ok {
		specifier, err := NewLibrarySpecifier(strings.Trim(variable.Client, "[]"))
		if err != nil {
			return VersionMeta{}, ForgeInstallProfile{}, fmt.Errorf("invalid official client specifier")
		}
		versionMeta.Libraries = append(versionMeta.Libraries, Library{
			Artifact: Artifact{
				Path: specifier.Path(),
			},
			Specifier:     specifier,
			ShouldInstall: true,
		})
	}
	versionMeta.LoaderID = version
	return versionMeta, profile, nil
}

// FetchPostProcessors retrieves arguments to run Forge's post processors for the specified game version.
func (forge forge) FetchPostProcessors(gameVersion, version string) ([]ForgeProcessor, error) {
	files, err := forge.FetchInstaller(version)
	if err != nil {
		return nil, fmt.Errorf("fetch installer: %w", err)
	}
	_, profile, err := forge.FetchMeta(version)
	if err != nil {
		return nil, fmt.Errorf("retrieve metadata: %w", err)
	}

	client, err := NewLibrarySpecifier(strings.Trim(profile.Data["PATCHED"].Client, "[]"))
	if err != nil {
		return nil, fmt.Errorf("invalid patched client specifier: %w", err)
	}
	if _, err := os.Stat(filepath.Join(env.LibrariesDir, client.Path())); err == nil {
		return nil, nil
	}

	var processors []processor
	variables := make(map[string]string)
	libraries := make(map[LibrarySpecifier]Library)

	for _, library := range profile.Libraries {
		libraries[library.Specifier] = library
	}

	for _, processor := range profile.Processors {
		if len(processor.Sides) > 0 && !slices.Contains(processor.Sides, "client") {
			continue
		}
		processors = append(processors, processor)
	}

	for k, v := range profile.Data {
		if strings.HasPrefix(v.Client, "/") {
			path := filepath.Join(env.TmpDir, v.Client)

			file, ok := files[v.Client[1:]]
			if !ok {
				return nil, fmt.Errorf("embedded installer file not present")
			}

			f, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("open embedded installer file: %w", err)
			}
			defer f.Close()

			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return nil, fmt.Errorf("create directory for file: %w", err)
			}

			df, err := os.Create(path)
			if err != nil {
				return nil, fmt.Errorf("create file: %w", err)
			}
			defer df.Close()
			if _, err := io.Copy(df, f); err != nil {
				return nil, fmt.Errorf("copy embedded installer file: %w", err)
			}
			v.Client = path
		}
		variables[k] = v.Client
	}

	variables["SIDE"] = "client"

	versionMeta, err := FetchVersionMeta(gameVersion)
	if err != nil {
		return nil, fmt.Errorf("retrieve version metadata: %w", err)
	}
	variables["MINECRAFT_JAR"] = versionMeta.Client().Artifact.RuntimePath()

	var post []ForgeProcessor

	for _, processor := range processors {
		jar, ok := libraries[processor.Jar]
		var mainClass string
		if !ok {
			return nil, fmt.Errorf("post processor library not found")
		}

		r, err := zip.OpenReader(jar.Artifact.RuntimePath())
		if err != nil {
			return nil, fmt.Errorf("read processor JAR: %w", err)
		}
		defer r.Close()
		files := make(map[string]*zip.File)
		for _, file := range r.File {
			files[file.Name] = file
		}

		file, ok := files["META-INF/MANIFEST.MF"]
		if !ok {
			return nil, fmt.Errorf("processor manifest not present")
		}
		f, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("open processor manifest: %w", err)
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "Main-Class: ") {
				mainClass = strings.TrimSpace(line[12:])
				break
			}
		}
		if mainClass == "" {
			return nil, fmt.Errorf("no main class found in processor manifest")
		}

		var paths []string
		for _, specifier := range processor.Classpath {
			library, ok := libraries[specifier]
			if !ok {
				return nil, fmt.Errorf("post processor library not found")
			}
			paths = append(paths, library.Artifact.RuntimePath())
		}
		var args []string
		for _, arg := range processor.Args {
			if arg[0] == '{' && arg[len(arg)-1] == '}' {
				arg = strings.Trim(arg, "{}")
				arg, ok = variables[arg]
				if !ok {
					return nil, fmt.Errorf("unknown processor argument")
				}
			}
			if arg[0] == '[' && arg[len(arg)-1] == ']' {
				arg = strings.Trim(arg, "[]")
				specifier, err := NewLibrarySpecifier(arg)
				if err != nil {
					return nil, fmt.Errorf("processor argument contains invalid library specifier")
				}
				arg = filepath.Join(env.LibrariesDir, specifier.Path())
			} else if arg[0] == '\'' && arg[len(arg)-1] == '\'' {
				arg = strings.Trim(arg, "'")
			}
			args = append(args, arg)
		}

		javaArgs := []string{
			"-cp", strings.Join(append(paths, jar.Artifact.RuntimePath()), string(os.PathListSeparator)),
			mainClass,
		}
		javaArgs = append(javaArgs, args...)

		post = append(post, ForgeProcessor{
			JavaArgs: javaArgs,
		})
	}
	return post, nil
}
