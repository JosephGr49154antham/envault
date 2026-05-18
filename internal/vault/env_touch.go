package vault

import (
	"fmt"
	"os"
	"time"
)

// TouchOptions configures the Touch operation.
type TouchOptions struct {
	// Keys is the list of keys to touch (upsert with empty value if missing).
	Keys []string
	// Dst is the output file path. Defaults to Src if empty.
	Dst string
	// Overwrite allows overwriting Dst when it differs from Src.
	Overwrite bool
}

// Touch ensures each key in opts.Keys exists in the env file.
// Keys that are already present are left unchanged; missing keys are
// appended with an empty value. The file's modification time is updated
// even when no keys are added.
func Touch(cfg Config, src string, opts TouchOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault is not initialised")
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("touch: read %s: %w", src, err)
	}

	lines := splitLines(string(data))
	existing := make(map[string]bool)
	for _, l := range lines {
		if k, _, ok := parseKV(l); ok {
			existing[k] = true
		}
	}

	for _, key := range opts.Keys {
		if !existing[key] {
			lines = append(lines, key+"=")
		}
	}

	dst := opts.Dst
	if dst == "" {
		dst = src
	}

	if dst != src {
		if _, err := os.Stat(dst); err == nil && !opts.Overwrite {
			return fmt.Errorf("touch: %s already exists; use overwrite flag", dst)
		}
	}

	out := joinLines(lines)
	if err := os.WriteFile(dst, []byte(out), 0o644); err != nil {
		return fmt.Errorf("touch: write %s: %w", dst, err)
	}

	now := time.Now()
	if err := os.Chtimes(dst, now, now); err != nil {
		return fmt.Errorf("touch: chtimes %s: %w", dst, err)
	}

	return nil
}

// splitLines splits s into lines, trimming a trailing newline.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	if len(s) > 0 && s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	out = append(out, s[start:])
	return out
}

// joinLines rejoins lines with newlines and appends a trailing newline.
func joinLines(lines []string) string {
	out := ""
	for _, l := range lines {
		out += l + "\n"
	}
	return out
}

// parseKV parses a KEY=VALUE line. Returns ok=false for comments/blanks.
func parseKV(line string) (key, value string, ok bool) {
	if len(line) == 0 || line[0] == '#' {
		return "", "", false
	}
	for i := 0; i < len(line); i++ {
		if line[i] == '=' {
			return line[:i], line[i+1:], true
		}
	}
	return "", "", false
}
