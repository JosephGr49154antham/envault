package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// TailOptions configures the Tail operation.
type TailOptions struct {
	Src      string
	N        int
	KeysOnly bool
}

// Tail returns the last N entries from an env file.
// Comments and blank lines are excluded from the count but preserved in output
// when KeysOnly is false.
func Tail(cfg Config, opts TailOptions) ([]string, error) {
	if !IsInitialised(cfg) {
		return nil, fmt.Errorf("vault not initialised: run envault init first")
	}

	src := opts.Src
	if src == "" {
		src = cfg.PlainFile
	}

	n := opts.N
	if n <= 0 {
		n = 10
	}

	f, err := os.Open(src)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", src, err)
	}
	defer f.Close()

	var entries []string
	var all []string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		all = append(all, line)
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		entries = append(entries, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", src, err)
	}

	if n > len(entries) {
		n = len(entries)
	}
	tail := entries[len(entries)-n:]

	if opts.KeysOnly {
		var keys []string
		for _, line := range tail {
			if idx := strings.Index(line, "="); idx > 0 {
				keys = append(keys, strings.TrimSpace(line[:idx]))
			} else {
				keys = append(keys, line)
			}
		}
		return keys, nil
	}

	_ = all
	return tail, nil
}
