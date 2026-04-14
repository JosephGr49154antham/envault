package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// StripOptions controls what Strip removes from an env file.
type StripOptions struct {
	// RemoveComments removes lines that are pure comments (start with #).
	RemoveComments bool
	// RemoveBlanks removes empty or whitespace-only lines.
	RemoveBlanks bool
	// Dst is the output path; if empty, src is overwritten in place.
	Dst string
}

// Strip reads src and writes a cleaned version to Dst (or src if Dst is empty),
// removing comments and/or blank lines according to opts.
func Strip(cfg Config, src string, opts StripOptions) (int, error) {
	if !IsInitialised(cfg) {
		return 0, fmt.Errorf("vault is not initialised; run 'envault init' first")
	}

	f, err := os.Open(src)
	if err != nil {
		return 0, fmt.Errorf("open %s: %w", src, err)
	}
	defer f.Close()

	var kept []string
	removed := 0

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if opts.RemoveComments && strings.HasPrefix(trimmed, "#") {
			removed++
			continue
		}
		if opts.RemoveBlanks && trimmed == "" {
			removed++
			continue
		}
		kept = append(kept, line)
	}
	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("reading %s: %w", src, err)
	}

	dst := opts.Dst
	if dst == "" {
		dst = src
	}

	out, err := os.Create(dst)
	if err != nil {
		return 0, fmt.Errorf("create %s: %w", dst, err)
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	for _, line := range kept {
		fmt.Fprintln(w, line)
	}
	if err := w.Flush(); err != nil {
		return 0, fmt.Errorf("write %s: %w", dst, err)
	}

	return removed, nil
}
