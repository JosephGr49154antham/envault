package vault

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Sort reads a .env file and writes it back with keys sorted alphabetically.
// Comments and blank lines are collected and emitted before the sorted key=value pairs.
// If dst is empty, the source file is sorted in place.
func Sort(cfg Config, src, dst string) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault is not initialised; run 'envault init' first")
	}

	if src == "" {
		src = cfg.PlainFile
	}
	if dst == "" {
		dst = src
	}

	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("sort: open %s: %w", src, err)
	}
	defer f.Close()

	var header []string // comments and blank lines at the top
	var pairs []string  // key=value lines

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			header = append(header, line)
		} else {
			pairs = append(pairs, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("sort: scan %s: %w", src, err)
	}

	sort.Slice(pairs, func(i, j int) bool {
		keyOf := func(s string) string {
			if idx := strings.IndexByte(s, '='); idx >= 0 {
				return strings.ToUpper(s[:idx])
			}
			return strings.ToUpper(s)
		}
		return keyOf(pairs[i]) < keyOf(pairs[j])
	})

	var sb strings.Builder
	for _, l := range header {
		sb.WriteString(l)
		sb.WriteByte('\n')
	}
	for _, l := range pairs {
		sb.WriteString(l)
		sb.WriteByte('\n')
	}

	if err := os.WriteFile(dst, []byte(sb.String()), 0o600); err != nil {
		return fmt.Errorf("sort: write %s: %w", dst, err)
	}
	return nil
}
