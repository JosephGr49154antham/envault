package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// TrimOptions controls which transformations Trim applies.
type TrimOptions struct {
	// RemoveComments strips comment lines (lines beginning with #).
	RemoveComments bool
	// RemoveBlanks strips blank / whitespace-only lines.
	RemoveBlanks bool
	// TrimValues removes leading and trailing whitespace from values.
	TrimValues bool
	// Dst is the output file path. Defaults to Src when empty.
	Dst string
}

// Trim reads the env file at src, applies the requested clean-up
// transformations and writes the result to opts.Dst (or back to src).
func Trim(cfg Config, src string, opts TrimOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault is not initialised – run 'envault init' first")
	}

	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("trim: open %s: %w", src, err)
	}
	defer f.Close()

	var out []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if opts.RemoveBlanks && trimmed == "" {
			continue
		}
		if opts.RemoveComments && strings.HasPrefix(trimmed, "#") {
			continue
		}
		if opts.TrimValues {
			line = trimEnvLine(line)
		}
		out = append(out, line)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("trim: scan %s: %w", src, err)
	}

	dst := opts.Dst
	if dst == "" {
		dst = src
	}

	if err := writeTrimLines(dst, out); err != nil {
		return fmt.Errorf("trim: write %s: %w", dst, err)
	}
	return nil
}

// trimEnvLine trims whitespace around the value part of a KEY=VALUE line.
func trimEnvLine(line string) string {
	if idx := strings.IndexByte(line, '='); idx >= 0 {
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		return key + "=" + val
	}
	return line
}

func writeTrimLines(path string, lines []string) error {
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
