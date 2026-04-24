package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ValuesResult holds the output of a Values operation.
type ValuesResult struct {
	Values []string
	File   string
}

// ValuesOptions controls the behaviour of Values.
type ValuesOptions struct {
	Src    string
	Keys   []string
	Quoted bool
}

// Values extracts the values (optionally for specific keys) from a .env file.
// If Keys is empty, all values are returned in file order.
func Values(cfg Config, opts ValuesOptions) (*ValuesResult, error) {
	if !IsInitialised(cfg) {
		return nil, fmt.Errorf("vault is not initialised: run 'envault init' first")
	}

	src := opts.Src
	if src == "" {
		src = cfg.PlainFile
	}

	f, err := os.Open(src)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", src, err)
	}
	defer f.Close()

	wantSet := make(map[string]bool, len(opts.Keys))
	for _, k := range opts.Keys {
		wantSet[strings.ToUpper(k)] = true
	}

	var values []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		idx := strings.IndexByte(trimmed, '=')
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(trimmed[:idx])
		val := strings.TrimSpace(trimmed[idx+1:])

		if len(wantSet) > 0 && !wantSet[strings.ToUpper(key)] {
			continue
		}

		if opts.Quoted {
			val = fmt.Sprintf("%q", val)
		}
		values = append(values, val)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", src, err)
	}

	return &ValuesResult{Values: values, File: src}, nil
}
