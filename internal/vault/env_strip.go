package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// StripOptions controls which lines are removed during stripping.
type StripOptions struct {
	// RemoveComments removes lines beginning with '#'.
	RemoveComments bool
	// RemoveBlanks removes empty or whitespace-only lines.
	RemoveBlanks bool
	// Dst is the output file path. If empty, the source file is overwritten.
	Dst string
}

// Strip reads an env file and writes a cleaned version to Dst (or in-place
// if Dst is empty), optionally removing comment lines and/or blank lines.
//
// The vault must be initialised before calling Strip.
func Strip(cfg Config, src string, opts StripOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault is not initialised; run 'envault init' first")
	}

	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer f.Close()

	var kept []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if opts.RemoveComments && strings.HasPrefix(trimmed, "#") {
			continue
		}
		if opts.RemoveBlanks && trimmed == "" {
			continue
		}
		kept = append(kept, line)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read source file: %w", err)
	}

	dst := opts.Dst
	if dst == "" {
		dst = src
	}

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	for _, line := range kept {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	}
	return w.Flush()
}
