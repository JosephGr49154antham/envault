package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// RenameValueOptions controls the behaviour of RenameValue.
type RenameValueOptions struct {
	Src      string
	Dst      string
	OldValue string
	NewValue string
	Overwrite bool
}

// RenameValue replaces all occurrences of OldValue with NewValue across the
// env file, writing the result to Dst. Only key=value lines are modified;
// comments and blank lines are preserved verbatim.
func RenameValue(cfg Config, opts RenameValueOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault not initialised: run 'envault init' first")
	}

	src := opts.Src
	if src == "" {
		src = ".env"
	}
	dst := opts.Dst
	if dst == "" {
		dst = src
	}

	if !opts.Overwrite && dst != src {
		if _, err := os.Stat(dst); err == nil {
			return fmt.Errorf("destination %q already exists; use --overwrite to replace", dst)
		}
	}

	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %q: %w", src, err)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			lines = append(lines, line)
			continue
		}
		if idx := strings.IndexByte(trimmed, '='); idx >= 0 {
			val := trimmed[idx+1:]
			unquoted := strings.Trim(val, `"`)
			if unquoted == opts.OldValue {
				key := trimmed[:idx]
				lines = append(lines, key+"="+opts.NewValue)
				continue
			}
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan %q: %w", src, err)
	}

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create %q: %w", dst, err)
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	for _, l := range lines {
		fmt.Fprintln(w, l)
	}
	return w.Flush()
}
