package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// UniqueResult holds the outcome of a Unique operation.
type UniqueResult struct {
	Removed  []string // keys that were duplicate and removed
	Kept     int      // number of lines kept
	OutputPath string
}

// Unique reads an env file, removes duplicate keys (keeping the last
// occurrence), and writes the result to dst. If dst is empty the source
// file is overwritten in-place.
func Unique(cfg Config, src, dst string) (UniqueResult, error) {
	if !IsInitialised(cfg) {
		return UniqueResult{}, fmt.Errorf("vault not initialised: run 'envault init' first")
	}

	if src == "" {
		src = ".env"
	}
	if dst == "" {
		dst = src
	}

	lines, err := readUniqueLines(src)
	if err != nil {
		return UniqueResult{}, fmt.Errorf("reading %s: %w", src, err)
	}

	deduped, removed := deduplicateLines(lines)

	if err := writeUniqueLines(dst, deduped); err != nil {
		return UniqueResult{}, fmt.Errorf("writing %s: %w", dst, err)
	}

	return UniqueResult{
		Removed:    removed,
		Kept:       len(deduped),
		OutputPath: dst,
	}, nil
}

func readUniqueLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// deduplicateLines keeps the last definition of each key.
// Comments and blank lines are preserved in their original relative positions,
// but when a key appears more than once all earlier occurrences are dropped.
func deduplicateLines(lines []string) (kept []string, removed []string) {
	// Two-pass approach: first build a map of key -> last line index,
	// then emit only lines that are either non-key or at their last index.
	lastIdx := map[string]int{}
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if idx := strings.IndexByte(trimmed, '='); idx > 0 {
			key := strings.TrimSpace(trimmed[:idx])
			lastIdx[key] = i
		}
	}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			kept = append(kept, line)
			continue
		}
		if idx := strings.IndexByte(trimmed, '='); idx > 0 {
			key := strings.TrimSpace(trimmed[:idx])
			if lastIdx[key] == i {
				kept = append(kept, line)
			} else {
				removed = append(removed, key)
			}
			continue
		}
		kept = append(kept, line)
	}
	return kept, removed
}

func writeUniqueLines(path string, lines []string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, l := range lines {
		fmt.Fprintln(w, l)
	}
	return w.Flush()
}
