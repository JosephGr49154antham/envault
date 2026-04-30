package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// SliceOptions controls the behaviour of the Slice operation.
type SliceOptions struct {
	Src   string
	Dst   string
	Start int // 1-based, inclusive
	End   int // 1-based, inclusive; 0 means EOF
	Force bool
}

// Slice extracts a range of key=value lines (by position, ignoring comments
// and blanks) from Src and writes them to Dst.
func Slice(cfg Config, opts SliceOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault not initialised: run 'envault init' first")
	}

	if opts.Src == "" {
		opts.Src = ".env"
	}
	if opts.Dst == "" {
		opts.Dst = opts.Src + ".slice"
	}
	if opts.Start < 1 {
		opts.Start = 1
	}

	f, err := os.Open(opts.Src)
	if err != nil {
		return fmt.Errorf("open %s: %w", opts.Src, err)
	}
	defer f.Close()

	var kvIndex int
	var out []string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		kvIndex++
		if kvIndex < opts.Start {
			continue
		}
		if opts.End > 0 && kvIndex > opts.End {
			break
		}
		out = append(out, line)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading %s: %w", opts.Src, err)
	}

	if !opts.Force {
		if _, err := os.Stat(opts.Dst); err == nil {
			return fmt.Errorf("destination %s already exists; use --force to overwrite", opts.Dst)
		}
	}

	dst, err := os.Create(opts.Dst)
	if err != nil {
		return fmt.Errorf("create %s: %w", opts.Dst, err)
	}
	defer dst.Close()

	w := bufio.NewWriter(dst)
	for _, l := range out {
		fmt.Fprintln(w, l)
	}
	return w.Flush()
}
