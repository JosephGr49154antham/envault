package vault

import (
	"os"
	"testing"
	"time"
)

func setupTagVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      dir + "/.envault",
		RecipientsFile: dir + "/.envault/recipients",
		EncryptedFile: dir + "/.env.age",
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return cfg, dir
}

func writeTestSnapshot(t *testing.T, cfg Config, id string) {
	t.Helper()
	dir := snapshotDir(cfg)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir snapshots: %v", err)
	}
	snap := Snapshot{
		ID:        id,
		Label:     "test",
		CreatedAt: time.Now().UTC(),
		Keys:      []string{"FOO"},
		Checksum:  "abc123",
	}
	path := snapshotDir(cfg) + "/" + id + ".json"
	data, _ := marshalSnapshot(snap)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write snapshot: %v", err)
	}
}

func TestCreateTag_Success(t *testing.T) {
	cfg, _ := setupTagVault(t)
	writeTestSnapshot(t, cfg, "snap-001")

	if err := CreateTag(cfg, "v1.0", "snap-001", "initial release"); err != nil {
		t.Fatalf("CreateTag: %v", err)
	}

	tags, err := ListTags(cfg)
	if err != nil {
		t.Fatalf("ListTags: %v", err)
	}
	if len(tags) != 1 {
		t.Fatalf("expected 1 tag, got %d", len(tags))
	}
	if tags[0].Name != "v1.0" || tags[0].SnapshotID != "snap-001" {
		t.Errorf("unexpected tag: %+v", tags[0])
	}
}

func TestCreateTag_SnapshotNotFound(t *testing.T) {
	cfg, _ := setupTagVault(t)
	err := CreateTag(cfg, "v1.0", "missing-snap", "")
	if err == nil {
		t.Fatal("expected error for missing snapshot")
	}
}

func TestCreateTag_NotInitialised(t *testing.T) {
	cfg := Config{VaultDir: t.TempDir() + "/.envault"}
	err := CreateTag(cfg, "v1", "snap-001", "")
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestListTags_Empty(t *testing.T) {
	cfg, _ := setupTagVault(t)
	tags, err := ListTags(cfg)
	if err != nil {
		t.Fatalf("ListTags: %v", err)
	}
	if len(tags) != 0 {
		t.Errorf("expected 0 tags, got %d", len(tags))
	}
}

func TestDeleteTag_Success(t *testing.T) {
	cfg, _ := setupTagVault(t)
	writeTestSnapshot(t, cfg, "snap-002")
	_ = CreateTag(cfg, "release", "snap-002", "")

	if err := DeleteTag(cfg, "release"); err != nil {
		t.Fatalf("DeleteTag: %v", err)
	}
	tags, _ := ListTags(cfg)
	if len(tags) != 0 {
		t.Errorf("expected 0 tags after delete, got %d", len(tags))
	}
}

func TestDeleteTag_NotFound(t *testing.T) {
	cfg, _ := setupTagVault(t)
	err := DeleteTag(cfg, "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing tag")
	}
}
