package vault

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
)

var nonAlphanumRe = regexp.MustCompile(`[^a-zA-Z0-9_-]`)

// sanitiseLabel makes a label safe for use in a filename.
func sanitiseLabel(label string) string {
	s := nonAlphanumRe.ReplaceAllString(label, "_")
	if len(s) > 40 {
		s = s[:40]
	}
	if s == "" {
		return "snapshot"
	}
	return s
}

// hashEnvValues returns a map of key -> hex(sha256(value)) so that values are
// never stored in plaintext in the snapshot file.
func hashEnvValues(env map[string]string) map[string]string {
	out := make(map[string]string, len(env))
	for k, v := range env {
		sum := sha256.Sum256([]byte(v))
		out[k] = fmt.Sprintf("%x", sum)
	}
	return out
}

// SnapshotDiff compares two snapshots and returns human-readable lines describing the diff.
func SnapshotDiff(a, b Snapshot) []string {
	var lines []string

	for k, hashA := range a.Keys {
		hashB, ok := b.Keys[k]
		if !ok {
			lines = append(lines, fmt.Sprintf("- %s (removed)", k))
		} else if hashA != hashB {
			lines = append(lines, fmt.Sprintf("~ %s (changed)", k))
		}
	}

	for k := range b.Keys {
		if _, ok := a.Keys[k]; !ok {
			lines = append(lines, fmt.Sprintf("+ %s (added)", k))
		}
	}

	if len(lines) == 0 {
		return []string{"  (no changes)"}
	}

	// stable sort for deterministic output
	sortStrings(lines)
	return lines
}

func sortStrings(ss []string) {
	for i := 1; i < len(ss); i++ {
		for j := i; j > 0 && strings.ToLower(ss[j]) < strings.ToLower(ss[j-1]); j-- {
			ss[j], ss[j-1] = ss[j-1], ss[j]
		}
	}
}
