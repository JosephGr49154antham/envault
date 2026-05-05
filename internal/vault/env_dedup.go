package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// DedupOptions controls the behaviour of the Dedup function.
type DedupOptions struct {
	// Src is the source .env file path.
	Src string
	// Dst is the destination path. If empty, Src is overwritten.
	Dst string
	// KeepLast retains the last occurrence of a duplicate key instead of the first.
	KeepLast bool
	// Overwrite allows writing to an existing Dst file.
	Overwrite bool
}

// Dedup reads a .env file and removes duplicate key definitions, keeping
// either the first or last occurrence depending on KeepLast.
func Dedup(cfg Config, opts DedupOptions) (int, error) {
	if !IsInitialised(cfg) {
		return 0, fmt.Errorf("vault not initialised: run 'envault init' first")
	}

	dst := opts.Dst
	if dst == "" {
		dst = opts.Src
	}

	if dst != opts.Src {
		if _, err := os.Stat(dst); err == nil && !opts.Overwrite {
			return 0, fmt.Errorf("destination file %q already exists; use --overwrite to replace", dst)
		}
	}

	lines, err := readDedupLines(opts.Src)
	if err != nil {
		return 0, err
	}

	deduped, removed := deduplicateEnvLines(lines, opts.KeepLast)

	out, err := os.Create(dst)
	if err != nil {
		return 0, fmt.Errorf("creating destination file: %w", err)
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	for _, l := range deduped {
		fmt.Fprintln(w, l)
	}
	if err := w.Flush(); err != nil {
		return 0, fmt.Errorf("writing deduplicated file: %w", err)
	}

	return removed, nil
}

func readDedupLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening source file %q: %w", path, err)
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}

func deduplicateEnvLines(lines []string, keepLast bool) ([]string, int) {
	type entry struct {
		index int
		line  string
	}

	keyIndex := map[string]int{} // key -> index in entries slice
	entries := make([]entry, 0, len(lines))
	removed := 0

	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			entries = append(entries, entry{index: len(entries), line: l})
			continue
		}

		key := envKey(trimmed)
		if key == "" {
			entries = append(entries, entry{index: len(entries), line: l})
			continue
		}

		if idx, seen := keyIndex[key]; seen {
			if keepLast {
				entries[idx].line = l
			}
			removed++
			continue
		}

		keyIndex[key] = len(entries)
		entries = append(entries, entry{index: len(entries), line: l})
	}

	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = e.line
	}
	return out, removed
}
