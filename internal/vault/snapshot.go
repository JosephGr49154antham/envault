package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Snapshot represents a point-in-time record of decrypted env key names and a hash of their values.
type Snapshot struct {
	CreatedAt time.Time         `json:"created_at"`
	Label     string            `json:"label"`
	Keys      map[string]string `json:"keys"` // key -> sha256 hex of value
}

func snapshotDir(cfg Config) string {
	return filepath.Join(cfg.VaultDir, "snapshots")
}

// SaveSnapshot encrypts the current .env file contents into a named snapshot stored in the vault.
func SaveSnapshot(cfg Config, label string) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault is not initialised; run 'envault init' first")
	}

	identity, err := loadVaultIdentity(cfg)
	if err != nil {
		return fmt.Errorf("load identity: %w", err)
	}

	envMap, err := decryptToMap(cfg, identity)
	if err != nil {
		return fmt.Errorf("decrypt env: %w", err)
	}

	hashedKeys := hashEnvValues(envMap)

	snap := Snapshot{
		CreatedAt: time.Now().UTC(),
		Label:     label,
		Keys:      hashedKeys,
	}

	if err := os.MkdirAll(snapshotDir(cfg), 0o700); err != nil {
		return fmt.Errorf("create snapshot dir: %w", err)
	}

	filename := fmt.Sprintf("%s_%s.json", snap.CreatedAt.Format("20060102T150405"), sanitiseLabel(label))
	path := filepath.Join(snapshotDir(cfg), filename)

	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write snapshot: %w", err)
	}

	return nil
}

// ListSnapshots returns all snapshots stored in the vault, ordered by filename (oldest first).
func ListSnapshots(cfg Config) ([]Snapshot, error) {
	if !IsInitialised(cfg) {
		return nil, fmt.Errorf("vault is not initialised; run 'envault init' first")
	}

	entries, err := os.ReadDir(snapshotDir(cfg))
	if os.IsNotExist(err) {
		return []Snapshot{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read snapshot dir: %w", err)
	}

	var snapshots []Snapshot
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(snapshotDir(cfg), e.Name()))
		if err != nil {
			return nil, fmt.Errorf("read snapshot %s: %w", e.Name(), err)
		}
		var snap Snapshot
		if err := json.Unmarshal(data, &snap); err != nil {
			return nil, fmt.Errorf("parse snapshot %s: %w", e.Name(), err)
		}
		snapshots = append(snapshots, snap)
	}
	return snapshots, nil
}
