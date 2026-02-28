package launcher

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/telecter/cmd-launcher/internal/meta"
)

// extractNatives extracts all native DLLs/SOs from legacy LWJGL 2 natives JARs
// into dest.
func extractNatives(dest string, libraries []meta.Library) error {
	for _, library := range libraries {
		for _, native := range library.Natives {
			path := native.Artifact.RuntimePath()

			if err := extractJar(path, dest); err != nil {
				return fmt.Errorf("extract natives from %s: %w", filepath.Base(path), err)
			}
		}
	}

	return nil
}

// extractJar extracts the contents of a ZIP/JAR at src into the directory dest,
// skipping files that already exist and META-INF entries.
func extractJar(src, dest string) error {
	var extractFile = func(f *zip.File, dest string) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = io.Copy(out, rc)
		return err
	}

	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	for _, f := range r.File {
		// Skip directories and metadata.
		if f.FileInfo().IsDir() {
			continue
		}
		if strings.HasPrefix(f.Name, "META-INF/") {
			continue
		}

		outPath := filepath.Join(dest, filepath.Base(f.Name))

		// Skip if already extracted.
		if _, err := os.Stat(outPath); err == nil {
			continue
		}

		if err := extractFile(f, outPath); err != nil {
			return err
		}
	}

	return nil
}
