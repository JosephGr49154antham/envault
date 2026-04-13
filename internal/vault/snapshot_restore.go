package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// RestoreSnapshot decrypts the snapshot identified by label and writes the
// plaintext .env content back to dst (defaults to cfg.PlainFile).
// The snapshot file must exist inside the vault snapshot directory.
func RestoreSnapshot(cfg Config, label, dst string) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault not initialised: run 'envault init' first")
	}

	dir := snapshotDir(cfg)
	name := sanitiseLabel(label) + ".json"
	snapshotPath := filepath.Join(dir, name)

	data, err := os.ReadFile(snapshotPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("snapshot %q not found", label)
		}
		return fmt.Errorf("reading snapshot: %w", err)
	}

	var snap struct {
		Label   string            `json:"label"`
		Entries map[string]string `json:"entries"`
	}
	if err := json.Unmarshal(data, &snap); err != nil {
		return fmt.Errorf("parsing snapshot: %w", err)
	}

	if dst == "" {
		dst = cfg.PlainFile
	}

	f, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("opening destination %q: %w", dst, err)
	}
	defer f.Close()

	for k, v := range snap.Entries {
		if _, err := fmt.Fprintf(f, "%s=%s\n", k, v); err != nil {
			return fmt.Errorf("writing entry: %w", err)
		}
	}

	return nil
}
