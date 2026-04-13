package vault_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rodrwan/envault/internal/vault"
)

// TestRedact_AllSensitivePatterns verifies every built-in pattern is caught.
func TestRedact_AllSensitivePatterns(t *testing.T) {
	cfg, dir := setupRedactVault(t)

	lines := []string{
		"MY_SECRET=v1",
		"DB_PASSWORD=v2",
		"DB_PASSWD=v3",
		"BEARER_TOKEN=v4",
		"APIKEY=v5",
		"API_KEY=v6",
		"PRIVATE_KEY=v7",
		"AWS_CREDENTIAL=v8",
		"AUTH_HEADER=v9",
		"JWT_SECRET=v10",
		"PASSPHRASE=v11",
		"SAFE_VALUE=keep",
	}
	src := writeRedactEnv(t, dir, ".env", strings.Join(lines, "\n")+"\n")
	dst := filepath.Join(dir, "out.env")

	result, err := vault.Redact(cfg, src, dst)
	if err != nil {
		t.Fatalf("Redact: %v", err)
	}

	if result.Total != 12 {
		t.Errorf("expected 12 total keys, got %d", result.Total)
	}
	if len(result.Redacted) != 11 {
		t.Errorf("expected 11 redacted keys, got %d: %v", len(result.Redacted), result.Redacted)
	}

	data, _ := os.ReadFile(dst)
	body := string(data)
	if !strings.Contains(body, "SAFE_VALUE=keep") {
		t.Error("expected SAFE_VALUE to be preserved")
	}
	for _, v := range []string{"v1", "v2", "v3", "v4", "v5", "v6", "v7", "v8", "v9", "v10", "v11"} {
		if strings.Contains(body, v) {
			t.Errorf("expected value %q to be redacted", v)
		}
	}
}

// TestRedact_EmptyFile handles an empty source gracefully.
func TestRedact_EmptyFile(t *testing.T) {
	cfg, dir := setupRedactVault(t)
	src := writeRedactEnv(t, dir, ".env", "")
	dst := filepath.Join(dir, "out.env")

	result, err := vault.Redact(cfg, src, dst)
	if err != nil {
		t.Fatalf("Redact: %v", err)
	}
	if result.Total != 0 || len(result.Redacted) != 0 {
		t.Errorf("expected zero keys for empty file, got total=%d redacted=%d", result.Total, len(result.Redacted))
	}
}
