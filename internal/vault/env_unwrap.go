package vault

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Unwrap reads a .env file where long values have been folded across multiple
// lines using a trailing backslash continuation and joins them back into single
// KEY=VALUE lines. Comments and blank lines are preserved as-is.
//
// Options:
//   - Src  – source .env file (required)
//   - Dst  – output path; defaults to <src>.unwrapped.env
//   - Overwrite – allow overwriting an existing destination file
type UnwrapOptions struct {
	Src       string
	Dst       string
	Overwrite bool
}

// Unwrap joins continuation lines in src and writes the result to dst.
func Unwrap(cfg Config, opts UnwrapOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault is not initialised; run 'envault init' first")
	}

	if opts.Src == "" {
		return fmt.Errorf("source file must be specified")
	}

	if opts.Dst == "" {
		ext := filepath.Ext(opts.Src)
		base := strings.TrimSuffix(opts.Src, ext)
		opts.Dst = base + ".unwrapped" + ext
	}

	if !opts.Overwrite {
		if _, err := os.Stat(opts.Dst); err == nil {
			return fmt.Errorf("destination file already exists: %s (use --overwrite to replace)", opts.Dst)
		}
	}

	lines, err := readUnwrapLines(opts.Src)
	if err != nil {
		return fmt.Errorf("reading source file: %w", err)
	}

	joined := joinContinuationLines(lines)

	if err := writeUnwrapLines(opts.Dst, joined); err != nil {
		return fmt.Errorf("writing destination file: %w", err)
	}

	return nil
}

// readUnwrapLines reads all raw lines from path.
func readUnwrapLines(path string) ([]string, error) {
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

// joinContinuationLines merges lines that end with a backslash with the
// following line. Comments (lines starting with #) and blank lines are
// emitted unchanged and never treated as continuations.
func joinContinuationLines(lines []string) []string {
	var out []string
	var buf strings.Builder
	inContinuation := false

	for _, line := range lines {
		trimmed := strings.TrimRight(line, " \t")

		// Comments and blank lines break any active continuation and are
		// emitted as-is.
		if strings.HasPrefix(trimmed, "#") || trimmed == "" {
			if inContinuation {
				out = append(out, buf.String())
				buf.Reset()
				inContinuation = false
			}
			out = append(out, line)
			continue
		}

		if strings.HasSuffix(trimmed, "\\") {
			// Strip the trailing backslash and accumulate.
			buf.WriteString(strings.TrimSuffix(trimmed, "\\"))
			inContinuation = true
		} else {
			buf.WriteString(trimmed)
			out = append(out, buf.String())
			buf.Reset()
			inContinuation = false
		}
	}

	// Flush any dangling continuation buffer.
	if buf.Len() > 0 {
		out = append(out, buf.String())
	}

	return out
}

// writeUnwrapLines writes lines to path, creating parent directories as needed.
func writeUnwrapLines(path string, lines []string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
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
