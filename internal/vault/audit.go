package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// AuditEvent represents a single recorded vault operation.
type AuditEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Operation string    `json:"operation"`
	User      string    `json:"user,omitempty"`
	Details   string    `json:"details,omitempty"`
}

// AuditLog holds a sequence of audit events.
type AuditLog struct {
	Events []AuditEvent `json:"events"`
}

// auditLogPath returns the path to the audit log file.
func auditLogPath(cfg Config) string {
	return filepath.Join(cfg.VaultDir, "audit.json")
}

// RecordEvent appends an audit event to the vault's audit log.
// If the vault is not initialised, it returns an error.
func RecordEvent(cfg Config, op, user, details string) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault not initialised")
	}

	log, err := LoadAuditLog(cfg)
	if err != nil {
		return fmt.Errorf("load audit log: %w", err)
	}

	log.Events = append(log.Events, AuditEvent{
		Timestamp: time.Now().UTC(),
		Operation: op,
		User:      user,
		Details:   details,
	})

	return saveAuditLog(cfg, log)
}

// LoadAuditLog reads and returns the audit log from disk.
// Returns an empty log if the file does not exist yet.
func LoadAuditLog(cfg Config) (AuditLog, error) {
	path := auditLogPath(cfg)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return AuditLog{}, nil
	}
	if err != nil {
		return AuditLog{}, fmt.Errorf("read audit log: %w", err)
	}

	var log AuditLog
	if err := json.Unmarshal(data, &log); err != nil {
		return AuditLog{}, fmt.Errorf("parse audit log: %w", err)
	}
	return log, nil
}

// saveAuditLog writes the audit log to disk as JSON.
func saveAuditLog(cfg Config, log AuditLog) error {
	data, err := json.MarshalIndent(log, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal audit log: %w", err)
	}
	if err := os.WriteFile(auditLogPath(cfg), data, 0o600); err != nil {
		return fmt.Errorf("write audit log: %w", err)
	}
	return nil
}
