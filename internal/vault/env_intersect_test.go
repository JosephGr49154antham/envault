package vault_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupIntersectVault(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := Init(DefaultConfig(dir)); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return dir
}

func writeIntersectEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func TestIntersect_CommonKeysOnly(t *testing.T) {
	dir := setupIntersectVault(t)

	src := filepath.Join(dir, ".env.a")
	other := filepath.Join(dir, ".env.b")
	dst := filepath.Join(dir, ".env.out")

	writeIntersectEnv(t, src, "FOO=1\nBAR=2\nBAZ=3\n")
	writeIntersectEnv(t, other, "BAR=99\nBAZ=88\nQUX=77\n")

	err := Intersect(IntersectOptions{
		VaultDir: dir,
		Src: src,
		Other: other,
		Dst: dst,
	})
	if err != nil {
		t.Fatalf("Intersect: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	out := string(data)

	if !strings.Contains(out, "BAR=") {
		t.Errorf("expected BAR in output, got:\n%s", out)
	}
	if !strings.Contains(out, "BAZ=") {
		t.Errorf("expected BAZ in output, got:\n%s", out)
	}
	if strings.Contains(out, "FOO=") {
		t.Errorf("FOO should not be in output (not in other), got:\n%s", out)
	}
	if strings.Contains(out, "QUX=") {
		t.Errorf("QUX should not be in output (not in src), got:\n%s", out)
	}
}

func TestIntersect_PrefersSrcValues(t *testing.T) {
	dir := setupIntersectVault(t)

	src := filepath.Join(dir, ".env.a")
	other := filepath.Join(dir, ".env.b")
	dst := filepath.Join(dir, ".env.out")

	writeIntersectEnv(t, src, "KEY=from_src\n")
	writeIntersectEnv(t, other, "KEY=from_other\n")

	if err := Intersect(IntersectOptions{VaultDir: dir, Src: src, Other: other, Dst: dst}); err != nil {
		t.Fatalf("Intersect: %v", err)
	}

	data, _ := os.ReadFile(dst)
	if !strings.Contains(string(data), "KEY=from_src") {
		t.Errorf("expected src value to be preserved, got: %s", string(data))
	}
}

func TestIntersect_DefaultDst(t *testing.T) {
	dir := setupIntersectVault(t)

	src := filepath.Join(dir, ".env")
	other := filepath.Join(dir, ".env.other")

	writeIntersectEnv(t, src, "A=1\nB=2\n")
	writeIntersectEnv(t, other, "B=9\nC=3\n")

	err := Intersect(IntersectOptions{VaultDir: dir, Src: src, Other: other})
	if err != nil {
		t.Fatalf("Intersect: %v", err)
	}

	expectedDst := filepath.Join(dir, ".env.intersect")
	if _, err := os.Stat(expectedDst); os.IsNotExist(err) {
		t.Errorf("expected default dst %s to exist", expectedDst)
	}
}

func TestIntersect_NoOverwrite(t *testing.T) {
	dir := setupIntersectVault(t)

	src := filepath.Join(dir, ".env.a")
	other := filepath.Join(dir, ".env.b")
	dst := filepath.Join(dir, ".env.out")

	writeIntersectEnv(t, src, "X=1\n")
	writeIntersectEnv(t, other, "X=2\n")
	writeIntersectEnv(t, dst, "EXISTING=yes\n")

	err := Intersect(IntersectOptions{VaultDir: dir, Src: src, Other: other, Dst: dst, Overwrite: false})
	if err == nil {
		t.Error("expected error when dst exists and overwrite=false")
	}
}

func TestIntersect_NotInitialised(t *testing.T) {
	dir := t.TempDir()

	err := Intersect(IntersectOptions{
		VaultDir: dir,
		Src: filepath.Join(dir, ".env.a"),
		Other: filepath.Join(dir, ".env.b"),
	})
	if err == nil {
		t.Error("expected error for uninitialised vault")
	}
}

func TestIntersect_EmptyIntersection(t *testing.T) {
	dir := setupIntersectVault(t)

	src := filepath.Join(dir, ".env.a")
	other := filepath.Join(dir, ".env.b")
	dst := filepath.Join(dir, ".env.out")

	writeIntersectEnv(t, src, "FOO=1\nBAR=2\n")
	writeIntersectEnv(t, other, "QUX=3\nZAP=4\n")

	if err := Intersect(IntersectOptions{VaultDir: dir, Src: src, Other: other, Dst: dst, Overwrite: true}); err != nil {
		t.Fatalf("Intersect: %v", err)
	}

	data, _ := os.ReadFile(dst)
	if strings.TrimSpace(string(data)) != "" {
		t.Errorf("expected empty output for disjoint files, got: %s", string(data))
	}
}
