package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupShuffleVault(t *testing.T) Config {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return cfg
}

func writeShuffleEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestShuffle_ReordersEntries(t *testing.T) {
	cfg := setupShuffleVault(t)
	src := filepath.Join(filepath.Dir(cfg.VaultDir), ".env")
	writeShuffleEnv(t, src, "A=1\nB=2\nC=3\nD=4\nE=5\n")

	opts := ShuffleOptions{Src: src, Dst: src, Seed: 42}
	if err := Shuffle(cfg, opts); err != nil {
		t.Fatalf("Shuffle: %v", err)
	}

	data, _ := os.ReadFile(src)
	lines := nonEmptyLines(string(data))
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines, got %d", len(lines))
	}
	// All original keys must still be present
	for _, key := range []string{"A", "B", "C", "D", "E"} {
		found := false
		for _, l := range lines {
			if strings.HasPrefix(l, key+"=") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("key %s missing after shuffle", key)
		}
	}
}

func TestShuffle_PreservesCommentsAndBlanks(t *testing.T) {
	cfg := setupShuffleVault(t)
	src := filepath.Join(filepath.Dir(cfg.VaultDir), ".env")
	writeShuffleEnv(t, src, "# header\nA=1\n\nB=2\n# mid\nC=3\n")

	opts := ShuffleOptions{Src: src, Dst: src, Seed: 7}
	if err := Shuffle(cfg, opts); err != nil {
		t.Fatalf("Shuffle: %v", err)
	}

	data, _ := os.ReadFile(src)
	content := string(data)
	if !strings.Contains(content, "# header") {
		t.Error("header comment lost")
	}
	if !strings.Contains(content, "# mid") {
		t.Error("mid comment lost")
	}
}

func TestShuffle_CustomDst(t *testing.T) {
	cfg := setupShuffleVault(t)
	dir := filepath.Dir(cfg.VaultDir)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.shuffled")
	writeShuffleEnv(t, src, "X=10\nY=20\n")

	if err := Shuffle(cfg, ShuffleOptions{Src: src, Dst: dst, Seed: 1}); err != nil {
		t.Fatalf("Shuffle: %v", err)
	}

	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("dst not created: %v", err)
	}
	// original unchanged
	origData, _ := os.ReadFile(src)
	if !strings.Contains(string(origData), "X=10") {
		t.Error("src was modified when dst differs")
	}
}

func TestShuffle_NotInitialised(t *testing.T) {
	cfg := Config{
		VaultDir: "/nonexistent/.envault",
	}
	err := Shuffle(cfg, ShuffleOptions{Src: ".env"})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

// nonEmptyLines returns lines that are not blank.
func nonEmptyLines(s string) []string {
	var out []string
	for _, l := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
		if strings.TrimSpace(l) != "" {
			out = append(out, l)
		}
	}
	return out
}
