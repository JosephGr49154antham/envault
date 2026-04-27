package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// LowerOptions configures the behaviour of the Lower operation.
type LowerOptions struct {
	// Keys restricts lowercasing to specific keys. If empty, all keys are lowercased.
	Keys []string
	// Dst is the output file path. If empty, the source file is overwritten.
	Dst string
	// Overwrite controls whether an existing Dst file may be replaced.
	Overwrite bool
}

// Lower reads an env file and writes a copy where the values of the specified
// keys (or all keys, if none are specified) are converted to lowercase.
// Keys and comments are preserved exactly as-is; only values are transformed.
func Lower(cfg Config, src string, opts LowerOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault not initialised: run 'envault init' first")
	}

	dst := opts.Dst
	if dst == "" {
		dst = src
	}

	if dst != src {
		if _, err := os.Stat(dst); err == nil && !opts.Overwrite {
			return fmt.Errorf("destination file already exists: %s (use --overwrite to replace)", dst)
		}
	}

	lines, err := readLowerLines(src)
	if err != nil {
		return fmt.Errorf("reading %s: %w", src, err)
	}

	keySet := make(map[string]bool, len(opts.Keys))
	for _, k := range opts.Keys {
		keySet[strings.ToUpper(k)] = true
	}

	var out []string
	for _, line := range lines {
		out = append(out, applyLower(line, keySet))
	}

	f, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("creating %s: %w", dst, err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, line := range out {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

// readLowerLines reads all lines from path, stripping the trailing newline
// that bufio.Scanner adds automatically.
func readLowerLines(path string) ([]string, error) {
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

// applyLower lowercases the value of a KEY=VALUE line when the key matches
// the target set (or when the set is empty, meaning all keys). Comment and
// blank lines are returned unchanged.
func applyLower(line string, keySet map[string]bool) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return line
	}

	eqIdx := strings.IndexByte(line, '=')
	if eqIdx < 0 {
		return line
	}

	key := strings.TrimSpace(line[:eqIdx])
	value := line[eqIdx+1:]

	if len(keySet) > 0 && !keySet[strings.ToUpper(key)] {
		return line
	}

	return key + "=" + strings.ToLower(value)
}
