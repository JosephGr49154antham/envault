package vault

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func setupWatchVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir: filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}
	src := filepath.Join(dir, ".env")
	if err := os.WriteFile(src, []byte("KEY=value\n"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return cfg, src
}

func TestWatch_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{VaultDir: filepath.Join(dir, ".envault")}
	stop := make(chan struct{})
	close(stop)
	err := Watch(cfg, ".env", stop, WatchOptions{})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestWatch_DetectsChange(t *testing.T) {
	cfg, src := setupWatchVault(t)

	var changed atomic.Int32
	stop := make(chan struct{})

	go func() {
		_ = Watch(cfg, src, stop, WatchOptions{
			Interval: 50 * time.Millisecond,
			OnChange: func(_ string) {
				changed.t	},
		})
	}()

	// Give the watcher time to record the baseline hash.
	time.Sleep(120 * time.Millisecond)

	// Modify the file.
	if err := os.WriteFile(src, []byte("KEY=new_value\n"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Wait for the watcher to detect the change.
	time.Sleep(200 * time.Millisecond)
	close(stop)

	if changed.Load() == 0 {
		t.Error("expected OnChange to be called at least once")
	}
}

func TestWatch_NoSpuriousFire(t *testing.T) {
	cfg, src := setupWatchVault(t)

	var changed atomic.Int32
	stop := make(chan struct{})

	go func() {
		_ = Watch(cfg, src, stop, WatchOptions{
			Interval: 50 * time.Millisecond,
			OnChange: func(_ string) { changed.Add(1) },
		})
	}()

	// File is never modified; wait several intervals.
	time.Sleep(300 * time.Millisecond)
	close(stop)

	if changed.Load() != 0 {
		t.Errorf("expected no OnChange calls, got %d", changed.Load())
	}
}

func TestWatch_OnError_CalledForMissingFile(t *testing.T) {
	cfg, _ := setupWatchVault(t)

	var errCount atomic.Int32
	stop := make(chan struct{})

	go func() {
		_ = Watch(cfg, "/nonexistent/.env", stop, WatchOptions{
			Interval: 50 * time.Millisecond,
			OnError:  func(_ string, _ error) { errCount.Add(1) },
		})
	}()

	time.Sleep(200 * time.Millisecond)
	close(stop)

	if errCount.Load() == 0 {
		t.Error("expected OnError to be called for missing file")
	}
}
