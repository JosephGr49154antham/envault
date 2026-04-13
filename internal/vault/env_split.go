package vault

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SplitOptions controls how an env file is split into multiple output files.
type SplitOptions struct {
	// Prefix maps a key prefix (e.g. "DB_") to an output file path.
	// Keys matching a prefix are written to the corresponding file.
	// Keys that match no prefix are written to the Remainder file (if set).
	Prefixes  map[string]string
	Remainder string // optional path for unmatched keys
}

// Split reads src and partitions its key=value pairs into separate files
// according to opts. Comment lines and blank lines are copied to every
// output file so context is preserved.
func Split(cfg Config, src string, opts SplitOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault is not initialised; run 'envault init' first")
	}

	if len(opts.Prefixes) == 0 && opts.Remainder == "" {
		return fmt.Errorf("split: no prefixes or remainder destination specified")
	}

	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("split: open %s: %w", src, err)
	}
	defer f.Close()

	// bucket: dest path -> lines
	buckets := make(map[string][]string)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// blank or comment — copy everywhere
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			for dest := range opts.Prefixes {
				buckets[dest] = append(buckets[dest], line)
			}
			if opts.Remainder != "" {
				buckets[opts.Remainder] = append(buckets[opts.Remainder], line)
			}
			continue
		}

		key := trimmed
		if idx := strings.IndexByte(trimmed, '='); idx >= 0 {
			key = trimmed[:idx]
		}

		matched := false
		for prefix, dest := range opts.Prefixes {
			if strings.HasPrefix(key, prefix) {
				buckets[dest] = append(buckets[dest], line)
				matched = true
				break
			}
		}
		if !matched && opts.Remainder != "" {
			buckets[opts.Remainder] = append(buckets[opts.Remainder], line)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("split: scan %s: %w", src, err)
	}

	for dest, lines := range buckets {
		if err := os.MkdirAll(filepath.Dir(dest), 0o700); err != nil {
			return fmt.Errorf("split: mkdir %s: %w", dest, err)
		}
		out := strings.Join(lines, "\n") + "\n"
		if err := os.WriteFile(dest, []byte(out), 0o600); err != nil {
			return fmt.Errorf("split: write %s: %w", dest, err)
		}
	}
	return nil
}
