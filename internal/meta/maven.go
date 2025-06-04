package meta

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/telecter/cmd-launcher/internal/network"
	env "github.com/telecter/cmd-launcher/pkg"
)

const MAVEN_REPO_URL = "https://repo.maven.apache.org/maven2/%s"

type LibrarySpecifier struct {
	Group      string
	Artifact   string
	Version    string
	Classifier string
}

func (specifier LibrarySpecifier) String() string {
	p := []string{specifier.Group, specifier.Artifact, specifier.Version}
	if specifier.Classifier != "" {
		p = append(p, specifier.Classifier)
	}
	return strings.Join(p, ":")
}
func (specifier LibrarySpecifier) Path() string {
	p := []string{specifier.Group, specifier.Artifact, specifier.Version}
	p[0] = strings.ReplaceAll(p[0], ".", "/")
	filename := p[1] + "-" + p[2]
	if specifier.Classifier != "" {
		filename += "-" + specifier.Classifier
	}
	filename += ".jar"
	return strings.Join([]string{p[0], p[1], p[2], filename}, "/")
}
func (specifier LibrarySpecifier) MarshalJSON() ([]byte, error) {
	return json.Marshal(specifier.String())
}
func (specifier *LibrarySpecifier) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	var err error
	*specifier, err = NewLibrarySpecifier(s)
	return err
}

func NewLibrarySpecifier(s string) (LibrarySpecifier, error) {
	p := strings.Split(s, ":")
	if len(p) < 3 {
		return LibrarySpecifier{}, fmt.Errorf("specifier too short")
	}
	specifier := LibrarySpecifier{
		Group:    p[0],
		Artifact: p[1],
		Version:  p[2],
	}
	if len(p) > 3 {
		specifier.Classifier = p[3]
	}
	return specifier, nil
}

// GetMavenLibrary returns library metadata for the specified name and path in the Maven repository.
func GetMavenLibrary(specifier LibrarySpecifier) (Library, error) {
	path := specifier.Path()
	url := fmt.Sprintf(MAVEN_REPO_URL, path)

	sumPath := filepath.Join(env.LibrariesDir, filepath.Dir(path), filepath.Base(path)+".sha1")
	var sum []byte
	sum, err := os.ReadFile(sumPath)

	if err != nil {
		resp, err := http.Get(url + ".sha1")
		if err != nil {
			return BaseLibrary{}, err
		}
		defer resp.Body.Close()
		if err := network.CheckResponse(resp); err != nil {
			return BaseLibrary{}, fmt.Errorf("library checksum could not fetched")
		}
		sum, _ = io.ReadAll(resp.Body)
		os.MkdirAll(filepath.Dir(sumPath), 0755)
		os.WriteFile(sumPath, sum, 0644)
	}
	return BaseLibrary{
		Name: specifier,
		LibraryArtifact: Artifact{
			Path: path,
			URL:  url,
			Sha1: string(sum),
		},
	}, nil
}
