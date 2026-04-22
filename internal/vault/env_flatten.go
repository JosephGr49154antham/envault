package vault

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FlattenOptions controls the behaviour of Flatten.
type FlattenOptions struct {
	// Prefix is prepended to every key (e.g. "APP_").
	Prefix string
	// Uppercase forces all keys to upper-case.
	Uppercase bool
	// Dst is the output file path. Defaults to <src>.flat.env.
	Dst string
	// Overwrite allows an existing destination file to be replaced.
	Overwrite bool
}

// Flatten reads a .env file and writes a normalised copy where every key
// is optionally prefixed and/or uppercased. Comments and blank lines are
// preserved as-is.
func Flatten(cfg Config, src string, opts FlattenOptions) (string, error) {
	if !IsInitialised(cfg) {
		return "", fmt.Errorf("vault is not initialised (run envault init)")
	}

	if _, err := os.Stat(src); err != nil {
		return "", fmt.Errorf("source file not found: %w", err)
	}

	dst := opts.Dst
	if dst == "" {
		ext := filepath.Ext(src)
		base := strings.TrimSuffix(src, ext)
		dst = base + ".flat" + ext
	}

	if !opts.Overwrite {
		if _, err := os.Stat(dst); err == nil {
			return "", fmt.Errorf("destination already exists: %s (use --overwrite to replace)", dst)
		}
	}

	lines, err := readFlattenLines(src, opts)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(dst, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		return "", fmt.Errorf("writing flattened file: %w", err)
	}

	return dst, nil
}

func readFlattenLines(src string, opts FlattenOptions) ([]string, error) {
	f, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			out = append(out, line)
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx < 0 {
			out = append(out, line)
			continue
		}
		key := line[:idx]
		val := line[idx+1:]
		if opts.Uppercase {
			key = strings.ToUpper(key)
		}
		if opts.Prefix != "" && !strings.HasPrefix(key, opts.Prefix) {
			key = opts.Prefix + key
		}
		out = append(out, key+"="+val)
	}
	return out, scanner.Err()
}
