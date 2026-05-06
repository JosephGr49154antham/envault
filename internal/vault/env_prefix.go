package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// PrefixOptions configures the Prefix operation.
type PrefixOptions struct {
	Src    string
	Dst    string
	Prefix string
	Remove bool // if true, strip the prefix instead of adding it
	Overwrite bool
}

// Prefix adds or removes a prefix from all keys in a .env file.
func Prefix(cfg Config, opts PrefixOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault is not initialised; run 'envault init' first")
	}
	if opts.Prefix == "" {
		return fmt.Errorf("prefix must not be empty")
	}
	if opts.Dst == "" {
		opts.Dst = opts.Src
	}
	if !opts.Overwrite && opts.Dst != opts.Src {
		if _, err := os.Stat(opts.Dst); err == nil {
			return fmt.Errorf("destination file already exists: %s", opts.Dst)
		}
	}

	lines, err := readPrefixLines(opts.Src)
	if err != nil {
		return fmt.Errorf("reading source file: %w", err)
	}

	out := applyPrefix(lines, opts.Prefix, opts.Remove)

	f, err := os.Create(opts.Dst)
	if err != nil {
		return fmt.Errorf("creating destination file: %w", err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, line := range out {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func readPrefixLines(path string) ([]string, error) {
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

func applyPrefix(lines []string, prefix string, remove bool) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			out = append(out, line)
			continue
		}
		eq := strings.IndexByte(trimmed, '=')
		if eq < 0 {
			out = append(out, line)
			continue
		}
		key := trimmed[:eq]
		val := trimmed[eq:]
		if remove {
			key = strings.TrimPrefix(key, prefix)
		} else {
			if !strings.HasPrefix(key, prefix) {
				key = prefix + key
			}
		}
		out = append(out, key+val)
	}
	return out
}
