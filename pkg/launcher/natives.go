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

// nativesDir returns the path to the natives extraction directory for an instance.
func nativesDir(instDir string) string {
	return filepath.Join(instDir, "natives")
}

// isLegacyNativesJar reports whether a library artifact path is a LWJGL2-style
// natives JAR that must be extracted manually (pre-1.14 / LWJGL 2).
func isLegacyNativesJar(artifactPath string) bool {
	base := filepath.Base(artifactPath)
	if !strings.Contains(base, "natives-") {
		return false
	}

	return strings.Contains(artifactPath, "lwjgl-platform") ||
		strings.Contains(artifactPath, "jinput-platform") ||
		strings.Contains(artifactPath, "org/lwjgl/lwjgl/")
}

// extractNatives extracts all native DLLs/SOs from legacy LWJGL 2 natives JARs
// into <instDir>/natives/.
func extractNatives(instDir string, libraries []meta.Library) error {
	dest := nativesDir(instDir)

	for _, lib := range libraries {
		path := lib.Artifact.RuntimePath()
		if !isLegacyNativesJar(path) {
			continue
		}

		if err := extractJar(path, dest); err != nil {
			return fmt.Errorf("extract natives from %s: %w", filepath.Base(path), err)
		}
	}

	return nil
}

// extractJar extracts the contents of a ZIP/JAR at src into the directory dest,
// skipping files that already exist and META-INF entries.
func extractJar(src, dest string) error {
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

// extractFile writes a single zip.File entry to outPath.
func extractFile(f *zip.File, outPath string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, rc)
	return err
}