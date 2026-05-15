package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// RenameKeyOptions controls the behaviour of RenameKey.
type RenameKeyOptions struct {
	Src     string
	Dst     string
	OldKey  string
	NewKey  string
	Force   bool // overwrite Dst if it already exists
}

// RenameKey reads Src, renames every occurrence of OldKey to NewKey,
// and writes the result to Dst (defaulting to Src when empty).
func RenameKey(cfg Config, opts RenameKeyOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault not initialised; run 'envault init' first")
	}

	src := opts.Src
	dst := opts.Dst
	if dst == "" {
		dst = src
	}

	if !opts.Force && dst != src {
		if _, err := os.Stat(dst); err == nil {
			return fmt.Errorf("destination file already exists: %s (use --force to overwrite)", dst)
		}
	}

	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer f.Close()

	var lines []string
	found := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			lines = append(lines, line)
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == opts.OldKey {
			lines = append(lines, opts.NewKey+"="+parts[1])
			found = true
		} else {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read source file: %w", err)
	}

	if !found {
		return fmt.Errorf("key %q not found in %s", opts.OldKey, src)
	}

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create destination file: %w", err)
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	for _, l := range lines {
		fmt.Fprintln(w, l)
	}
	return w.Flush()
}
