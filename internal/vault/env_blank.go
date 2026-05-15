package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// BlankOptions controls the behaviour of the Blank operation.
type BlankOptions struct {
	// Keys is the list of keys whose values should be blanked.
	// If empty, all key=value pairs are blanked.
	Keys []string
	// Dst is the output file path. Defaults to Src if empty.
	Dst string
	// Overwrite allows overwriting an existing Dst file.
	Overwrite bool
}

// Blank reads an env file and writes a copy where selected (or all) values
// are replaced with empty strings, preserving comments and blank lines.
func Blank(cfg Config, src string, opts BlankOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault not initialised")
	}

	dst := opts.Dst
	if dst == "" {
		dst = src
	}

	if !opts.Overwrite && dst != src {
		if _, err := os.Stat(dst); err == nil {
			return fmt.Errorf("destination file already exists: %s", dst)
		}
	}

	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer f.Close()

	keySet := make(map[string]bool, len(opts.Keys))
	for _, k := range opts.Keys {
		keySet[strings.TrimSpace(k)] = true
	}

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			lines = append(lines, line)
			continue
		}
		idx := strings.IndexByte(trimmed, '=')
		if idx < 0 {
			lines = append(lines, line)
			continue
		}
		key := strings.TrimSpace(trimmed[:idx])
		if len(keySet) == 0 || keySet[key] {
			lines = append(lines, key+"=")
		} else {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan: %w", err)
	}

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create destination: %w", err)
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	for _, l := range lines {
		fmt.Fprintln(w, l)
	}
	return w.Flush()
}
