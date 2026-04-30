package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupSliceVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}
	return cfg, dir
}

func writeSliceEnv(t *testing.T, dir, content string) string {
	t.Helper()
	p := filepath.Join(dir, ".env")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
	return p
}

func TestSlice_ExtractsRange(t *testing.T) {
	cfg, dir := setupSliceVault(t)
	src := writeSliceEnv(t, dir, "A=1\nB=2\nC=3\nD=4\nE=5\n")
	dst := filepath.Join(dir, "out.env")

	err := Slice(cfg, SliceOptions{Src: src, Dst: dst, Start: 2, End: 4, Force: false})
	if err != nil {
		t.Fatalf("Slice: %v", err)
	}

	data, _ := os.ReadFile(dst)
	got := strings.TrimSpace(string(data))
	want := "B=2\nC=3\nD=4"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSlice_SkipsCommentsAndBlanks(t *testing.T) {
	cfg, dir := setupSliceVault(t)
	src := writeSliceEnv(t, dir, "# header\nA=1\n\nB=2\nC=3\n")
	dst := filepath.Join(dir, "out.env")

	// kv positions: A=1 -> 1, B=2 -> 2, C=3 -> 3
	err := Slice(cfg, SliceOptions{Src: src, Dst: dst, Start: 2, End: 3, Force: false})
	if err != nil {
		t.Fatalf("Slice: %v", err)
	}

	data, _ := os.ReadFile(dst)
	got := strings.TrimSpace(string(data))
	if !strings.Contains(got, "B=2") || !strings.Contains(got, "C=3") {
		t.Errorf("unexpected output: %q", got)
	}
	if strings.Contains(got, "A=1") {
		t.Errorf("A=1 should not be in slice output")
	}
}

func TestSlice_NoOverwrite(t *testing.T) {
	cfg, dir := setupSliceVault(t)
	src := writeSliceEnv(t, dir, "A=1\nB=2\n")
	dst := filepath.Join(dir, "out.env")
	os.WriteFile(dst, []byte("existing"), 0o600)

	err := Slice(cfg, SliceOptions{Src: src, Dst: dst, Start: 1, End: 1, Force: false})
	if err == nil {
		t.Fatal("expected error when dst exists and force=false")
	}
}

func TestSlice_ForceOverwrite(t *testing.T) {
	cfg, dir := setupSliceVault(t)
	src := writeSliceEnv(t, dir, "A=1\nB=2\n")
	dst := filepath.Join(dir, "out.env")
	os.WriteFile(dst, []byte("old"), 0o600)

	err := Slice(cfg, SliceOptions{Src: src, Dst: dst, Start: 1, End: 1, Force: true})
	if err != nil {
		t.Fatalf("Slice with force: %v", err)
	}
	data, _ := os.ReadFile(dst)
	if !strings.Contains(string(data), "A=1") {
		t.Errorf("expected A=1 in output, got %q", string(data))
	}
}

func TestSlice_NotInitialised(t *testing.T) {
	cfg := Config{VaultDir: t.TempDir() + "/missing"}
	err := Slice(cfg, SliceOptions{Src: ".env", Dst: "out.env", Start: 1})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}
