package vault

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WrapOptions controls how values are wrapped.
type WrapOptions struct {
	Src       string
	Dst       string
	Width     int
	Overwrite bool
}

// Wrap folds long env values at the given column width using shell line
// continuations (backslash-newline). Comments and blank lines are preserved
// unchanged.
func Wrap(cfg Config, opts WrapOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault not initialised; run 'envault init' first")
	}
	if opts.Width <= 0 {
		opts.Width = 80
	}
	if opts.Src == "" {
		opts.Src = ".env"
	}
	if opts.Dst == "" {
		opts.Dst = opts.Src
	}

	lines, err := readWrapLines(opts.Src)
	if err != nil {
		return err
	}

	if !opts.Overwrite {
		if _, err := os.Stat(opts.Dst); err == nil && opts.Dst != opts.Src {
			return fmt.Errorf("destination %q already exists; use --overwrite to replace", opts.Dst)
		}
	}

	if err := os.MkdirAll(filepath.Dir(opts.Dst), 0o755); err != nil {
		return err
	}

	out, err := os.Create(opts.Dst)
	if err != nil {
		return err
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func readWrapLines(src string) ([]string, error) {
	f, err := os.Open(src)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", src, err)
	}
	defer f.Close()

	var out []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		out = append(out, scanner.Text())
	}
	return out, scanner.Err()
}

// wrapValue inserts backslash-newline continuations so that the KEY=VALUE line
// does not exceed width characters.
func wrapValue(key, value string, width int) []string {
	prefix := key + "="
	full := prefix + value
	if len(full) <= width {
		return []string{full}
	}

	var lines []string
	remaining := value
	first := true
	for len(remaining) > 0 {
		var pfx string
		if first {
			pfx = prefix
			first = false
		} else {
			pfx = strings.Repeat(" ", len(prefix))
		}
		avail := width - len(pfx) - 1 // -1 for trailing \
		if avail <= 0 {
			avail = 1
		}
		if len(remaining) <= avail {
			lines = append(lines, pfx+remaining)
			break
		}
		lines = append(lines, pfx+remaining[:avail]+"\\")
		remaining = remaining[avail:]
	}
	return lines
}
