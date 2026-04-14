package vault

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"time"
)

// WatchOptions configures the Watch behaviour.
type WatchOptions struct {
	// Interval between polls. Defaults to 2 seconds.
	Interval time.Duration
	// OnChange is called whenever the file content changes.
	OnChange func(path string)
	// OnError is called when the file cannot be read.
	OnError func(path string, err error)
}

// Watch polls a plain .env file for changes and invokes opts.OnChange when
// the file's SHA-256 digest differs from the previous read. It blocks until
// ctx is done (pass a context.Context-compatible done channel via stop).
//
// The vault must be initialised before calling Watch.
func Watch(cfg Config, src string, stop <-chan struct{}, opts WatchOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault not initialised at %s", cfg.VaultDir)
	}
	if opts.Interval <= 0 {
		opts.Interval = 2 * time.Second
	}
	if opts.OnChange == nil {
		opts.OnChange = func(string) {}
	}
	if opts.OnError == nil {
		opts.OnError = func(string, error) {}
	}

	var lastHash [32]byte
	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return nil
		case <-ticker.C:
			h, err := fileHash(src)
			if err != nil {
				opts.OnError(src, err)
				continue
			}
			if h != lastHash {
				if lastHash != ([32]byte{}) {
					opts.OnChange(src)
				}
				lastHash = h
			}
		}
	}
}

func fileHash(path string) ([32]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return [32]byte{}, err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return [32]byte{}, err
	}
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out, nil
}
