package launcher

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/telecter/cmd-launcher/internal/meta"
)

// filterLibraries sorts game libraries into installed and required libraries.
//
// It also provides fixes and swaps out some libraries to ensure compatability.
func filterLibraries(libraries []meta.Library) (installed []meta.Library, required []meta.Library) {
	m := make(map[string]int)
	for _, library := range libraries {
		l := []meta.Library{library}

		k := strings.Join([]string{library.Specifier().Group, library.Specifier().Artifact, library.Specifier().Classifier}, ":")
		switch library := library.(type) {
		case meta.MojangLibrary:
			if m[k] > 0 {
				continue
			}
			l = append(l, library.Classifiers()...)
		default:
			m[k]++
		}

		if !library.ShouldInstall() {
			continue
		}

		for _, library := range l {
			if library.Artifact().URL == "" {
				continue
			}
			library = patchLibrary(library)
			if library.Artifact().IsDownloaded() {
				installed = append(installed, library)
			} else {
				required = append(required, library)
			}
		}
	}
	return installed, required
}

// patchLibrary takes library and applies any applicable fixes to it.
func patchLibrary(library meta.Library) meta.Library {
	specifier := library.Specifier()

	if specifier.Group == "org.lwjgl" &&
		specifier.Classifier == "natives-linux" &&
		runtime.GOOS == "linux" &&
		runtime.GOARCH == "arm64" {
		v, err := semver.NewVersion(specifier.Version)
		if err != nil {
			return library
		}
		if specifier.Artifact == "lwjgl-jemalloc" && v.LessThan(semver.MustParse("3.3.2")) && v.GreaterThanEqual(semver.MustParse("3.0.0")) {
			return meta.BaseLibrary{
				Name: specifier,
				LibraryArtifact: meta.Artifact{
					URL:  fmt.Sprintf("https://raw.githubusercontent.com/theofficialgman/lwjgl3-binaries-arm64/refs/heads/lwjgl-%s/lwjgl-jemalloc-patched-natives-linux-arm64.jar", v.String()),
					Path: library.Artifact().Path,
				},
			}
		}
		c, err := semver.NewConstraint("3.1.6 - 3.2.2 != 3.2.0")
		if err != nil {
			panic(err)
		}
		if c.Check(v) {
			return meta.BaseLibrary{
				Name: specifier,
				LibraryArtifact: meta.Artifact{
					URL:  fmt.Sprintf("https://github.com/theofficialgman/lwjgl3-binaries-arm64/raw/refs/heads/lwjgl-combined/%s", library.Artifact().Path),
					Path: library.Artifact().Path,
				},
			}
		}

		specifier.Classifier = "natives-linux-arm64"
		library, err := meta.GetMavenLibrary(specifier)
		if err == nil {
			return library
		}
	}

	return library
}
