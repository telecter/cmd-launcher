package meta

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/telecter/cmd-launcher/internal/network"
	env "github.com/telecter/cmd-launcher/pkg"
)

const MavenRepoURL = "https://repo.maven.apache.org/maven2"

// A LibrarySpecifier represents the Maven specifier syntax.
//
// group:artifact:version:classifier
type LibrarySpecifier struct {
	Group      string
	Artifact   string
	Version    string
	Classifier string
}

// String returns the string form of the specifier.
//
// It is formatted as: group:artifact:version:classifier
func (specifier LibrarySpecifier) String() string {
	p := []string{specifier.Group, specifier.Artifact, specifier.Version}
	if specifier.Classifier != "" {
		p = append(p, specifier.Classifier)
	}
	return strings.Join(p, ":")
}

// Path returns the relative path to the JAR file described by the specifier.
func (specifier LibrarySpecifier) Path() string {
	t := ".jar"
	if strings.HasSuffix(specifier.Version, "@zip") {
		specifier.Version = strings.ReplaceAll(specifier.Version, "@zip", "")
		t = ".zip"
	}

	p := []string{specifier.Group, specifier.Artifact, specifier.Version}

	p[0] = strings.ReplaceAll(p[0], ".", "/")
	filename := p[1] + "-" + p[2]
	if specifier.Classifier != "" {
		filename += "-" + specifier.Classifier
	}
	filename += t
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

// NewLibrarySpecifier creates a new LibrarySpecifier from its string form.
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

// FetchMavenLibrary returns library metadata for the specified name and path in the Maven repository.
func FetchMavenLibrary(specifier LibrarySpecifier) (Library, error) {
	url, _ := url.JoinPath(MavenRepoURL, specifier.Path())
	path := specifier.Path()

	sumPath := filepath.Join(env.LibrariesDir, filepath.Dir(path), filepath.Base(path)+".sha1")
	var sum []byte
	sum, err := os.ReadFile(sumPath)

	if err != nil {
		resp, err := http.Get(url + ".sha1")
		if err != nil {
			return Library{}, err
		}
		defer resp.Body.Close()
		if err := network.CheckResponse(resp); err != nil {
			return Library{}, fmt.Errorf("retrieve library checksum: %w", err)
		}
		sum, _ = io.ReadAll(resp.Body)
		os.MkdirAll(filepath.Dir(sumPath), 0755)
		os.WriteFile(sumPath, sum, 0644)
	}
	return Library{
		Specifier: specifier,
		Artifact: Artifact{
			Path: path,
			URL:  url,
			Sha1: string(sum),
		},
	}, nil
}
