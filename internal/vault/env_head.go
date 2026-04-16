package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// HeadOptions controls the behaviour of Head.
type HeadOptions struct {
	// N is the number of lines to display (default 10).
	N int
	// KeysOnly omits values, printing only key names.
	KeysOnly bool
	// Dst is an optional output path; if empty, results are returned as a string.
	Dst string
}

// Head returns the first N lines of a .env file, skipping blank lines and
// comments when KeysOnly is true. If Dst is set the output is written there;
// otherwise the rendered text is returned.
func Head(cfg Config, src string, opts HeadOptions) (string, error) {
	if !IsInitialised(cfg) {
		return "", fmt.Errorf("vault not initialised: run 'envault init' first")
	}

	if opts.N <= 0 {
		opts.N = 10
	}

	f, err := os.Open(src)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", src, err)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() && len(lines) < opts.N {
		line := scanner.Text()
		if opts.KeysOnly {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			key, _, _ := strings.Cut(trimmed, "=")
			lines = append(lines, strings.TrimSpace(key))
		} else {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read %s: %w", src, err)
	}

	out := strings.Join(lines, "\n")
	if out != "" {
		out += "\n"
	}

	if opts.Dst != "" {
		if err := os.WriteFile(opts.Dst, []byte(out), 0o600); err != nil {
			return "", fmt.Errorf("write %s: %w", opts.Dst, err)
		}
		return opts.Dst, nil
	}
	return out, nil
}
