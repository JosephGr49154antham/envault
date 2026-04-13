package vault

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Tag represents a named marker pointing.
type Tag struct {
	Name       string    `json:"name"`
	SnapshotID string    `json:"snapshot_id"`
	Message    string    `json:"message,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

func tagDir(cfg Config) string {
	return filepath.Join(cfg.VaultDir, "tags")
}

func tagPath(cfg Config, name string) string {
	return filepath.Join(tagDir(cfg), sanitiseLabel(name)+".json")
}

// Create tag pointing to an existing snapshot ID.
func CreateTag(cfg Config, name, snapshotID, message string) error {
	if !IsInitialised(cfg) {
		return errors.New("vault is not initialised")
	}
	if name == "" {
		return errors.New("tag name must not be empty")
	}
	if snapshotID == "" {
		return errors.New("snapshot ID must not be empty")
	}

	// Verify the snapshot exists.
	snaps, err := ListSnapshots(cfg)
	if err != nil {
		return fmt.Errorf("listing snapshots: %w", err)
	}
	found := false
	for _, s := range snaps {
		if s.ID == snapshotID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("snapshot %q not found", snapshotID)
	}

	if err := os.MkdirAll(tagDir(cfg), 0o700); err != nil {
		return fmt.Errorf("creating tag dir: %w", err)
	}

	tag := Tag{
		Name:       name,
		SnapshotID: snapshotID,
		Message:    message,
		CreatedAt:  time.Now().UTC(),
	}
	data, err := json.MarshalIndent(tag, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(tagPath(cfg, name), data, 0o600)
}

// ListTags returns all tags stored in the vault.
func ListTags(cfg Config) ([]Tag, error) {
	if !IsInitialised(cfg) {
		return nil, errors.New("vault is not initialised")
	}
	entries, err := os.ReadDir(tagDir(cfg))
	if os.IsNotExist(err) {
		return []Tag{}, nil
	}
	if err != nil {
		return nil, err
	}
	var tags []Tag
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(tagDir(cfg), e.Name()))
		if err != nil {
			return nil, err
		}
		var t Tag
		if err := json.Unmarshal(data, &t); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, nil
}

// DeleteTag removes a named tag from the vault.
func DeleteTag(cfg Config, name string) error {
	if !IsInitialised(cfg) {
		return errors.New("vault is not initialised")
	}
	p := tagPath(cfg, name)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return fmt.Errorf("tag %q not found", name)
	}
	return os.Remove(p)
}
