package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// CompactOptions controls the behaviour of Compact.
type CompactOptions struct {
	Src       string
	Dst       string
	Overwrite bool
}

// Compact removes blank lines and comment-only lines from an env file,
// writing the result to Dst. If Dst is empty it defaults to Src.
func Compact(cfg Config, opts CompactOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault is not initialised; run 'envault init' first")
	}

	if opts.Dst == "" {
		opts.Dst = opts.Src
	}

	if !opts.Overwrite && opts.Dst != opts.Src {
		if _, err := os.Stat(opts.Dst); err == nil {
			return fmt.Errorf("destination already exists: %s (use --overwrite to replace)", opts.Dst)
		}
	}

	lines, err := readCompactLines(opts.Src)
	if err != nil {
		return err
	}

	var kept []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		kept = append(kept, line)
	}

	return writeCompactLines(opts.Dst, kept)
}

func readCompactLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("compact: open %s: %w", path, err)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func writeCompactLines(path string, lines []string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("compact: create %s: %w", path, err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}
