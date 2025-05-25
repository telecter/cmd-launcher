package meta

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/telecter/cmd-launcher/internal"
)

const MAVEN_REPO_URL = "https://repo.maven.apache.org/maven2/%s"

func GetMavenLibrary(name string, path string) Library {
	url := fmt.Sprintf(MAVEN_REPO_URL, path)
	sumPath := filepath.Join(internal.LibrariesDir, filepath.Dir(path), filepath.Base(path)+".sha1")
	var sum []byte
	sum, err := os.ReadFile(sumPath)
	if err != nil {
		resp, err := http.Get(url + ".sha1")
		if err == nil {
			defer resp.Body.Close()
			sum, _ = io.ReadAll(resp.Body)

			os.MkdirAll(filepath.Dir(sumPath), 0755)
			os.WriteFile(sumPath, sum, 0644)
		} else {
			sum = []byte{}
		}
	}
	return Library{
		Name: name,
		Downloads: struct {
			Artifact Artifact "json:\"artifact\""
		}{
			Artifact: Artifact{
				Path: path,
				URL:  url,
				Sha1: string(sum),
			},
		},
	}
}
