package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// AppendOptions controls the behaviour of the Append operation.
type AppendOptions struct {
	Src     string
	Dst     string
	Keys    []string // if non-empty, only append these keys
	NoBlank bool     // skip blank lines from src
	DryRun  bool
}

// Append reads key=value pairs from Src and appends them to Dst,
// skipping any key that already exists in Dst.
func Append(cfg Config, opts AppendOptions) ([]string, error) {
	if !IsInitialised(cfg) {
		return nil, fmt.Errorf("vault not initialised: run 'envault init' first")
	}

	srcLines, err := readAppendLines(opts.Src)
	if err != nil {
		return nil, fmt.Errorf("reading source file: %w", err)
	}

	existing, err := loadEnvKeysFromFile(opts.Dst)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("reading destination file: %w", err)
	}

	filter := make(map[string]bool, len(opts.Keys))
	for _, k := range opts.Keys {
		filter[k] = true
	}

	var toAppend []string
	var appended []string
	for _, line := range srcLines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if !opts.NoBlank {
				toAppend = append(toAppend, line)
			}
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			toAppend = append(toAppend, line)
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		if len(filter) > 0 && !filter[key] {
			continue
		}
		if existing[key] {
			continue
		}
		toAppend = append(toAppend, line)
		appended = append(appended, key)
	}

	if opts.DryRun || len(appended) == 0 {
		return appended, nil
	}

	f, err := os.OpenFile(opts.Dst, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("opening destination file: %w", err)
	}
	defer f.Close()

	for _, line := range toAppend {
		if _, err := fmt.Fprintln(f, line); err != nil {
			return nil, fmt.Errorf("writing to destination: %w", err)
		}
	}
	return appended, nil
}

func readAppendLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}
