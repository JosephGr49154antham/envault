package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInterpolate_DollarSignWithoutBraces(t *testing.T) {
	cfg, dir := setupInterpolateVault(t)
	src := filepath.Join(dir, ".env")
	writeInterpolateEnv(t, src, "DIR=/data\nFILE=$DIR/config.json\n")

	dst := filepath.Join(dir, "out.env")
	if err := Interpolate(cfg, InterpolateOptions{Src: src, Dst: dst}); err != nil {
		t.Fatalf("Interpolate: %v", err)
	}
	out, _ := os.ReadFile(dst)
	if !strings.Contains(string(out), "FILE=/data/config.json") {
		t.Errorf("expected $VAR expansion, got:\n%s", out)
	}
}

func TestInterpolate_FallsBackToOSEnv(t *testing.T) {
	cfg, dir := setupInterpolateVault(t)
	t.Setenv("ENVAULT_TEST_HOST", "os-host.example.com")

	src := filepath.Join(dir, ".env")
	writeInterpolateEnv(t, src, "URL=https://${ENVAULT_TEST_HOST}/api\n")

	dst := filepath.Join(dir, "out.env")
	if err := Interpolate(cfg, InterpolateOptions{Src: src, Dst: dst}); err != nil {
		t.Fatalf("Interpolate: %v", err)
	}
	out, _ := os.ReadFile(dst)
	if !strings.Contains(string(out), "URL=https://os-host.example.com/api") {
		t.Errorf("expected OS env fallback, got:\n%s", out)
	}
}

func TestInterpolate_DefaultDstPath(t *testing.T) {
	cfg, dir := setupInterpolateVault(t)
	src := filepath.Join(dir, ".env")
	writeInterpolateEnv(t, src, "A=1\n")

	if err := Interpolate(cfg, InterpolateOptions{Src: src}); err != nil {
		t.Fatalf("Interpolate: %v", err)
	}

	expected := filepath.Join(dir, ".interpolated.env")
	if _, err := os.Stat(expected); os.IsNotExist(err) {
		// Try alternative default naming
		alt := filepath.Join(dir, ".env.interpolated")
		if _, err2 := os.Stat(alt); os.IsNotExist(err2) {
			// Accept any file with "interpolated" in name in dir
			entries, _ := os.ReadDir(dir)
			found := false
			for _, e := range entries {
				if strings.Contains(e.Name(), "interpolated") {
					found = true
				}
			}
			if !found {
				t.Errorf("expected an interpolated output file in %s", dir)
			}
		}
	}
}
