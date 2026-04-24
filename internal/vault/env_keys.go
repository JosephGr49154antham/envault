package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// KeysOptions controls the output behaviour of Keys.
type KeysOptions struct {
	Sorted bool
	ValuesOnly bool
}

// Keys extracts all variable names (or values) from a .env file.
// If cfg.Sorted is true the output is sorted alphabetically.
// If cfg.ValuesOnly is true the values are returned instead of keys.
func Keys(cfg Config, src string, opts KeysOptions) ([]string, error) {
	if !IsInitialised(cfg) {
		return nil, fmt.Errorf("vault not initialised; run 'envault init' first")
	}

	f, err := os.Open(src)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", src, err)
	}
	defer f.Close()

	var results []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx < 1 {
			continue
		}
		if opts.ValuesOnly {
			results = append(results, strings.TrimSpace(line[idx+1:]))
		} else {
			results = append(results, strings.TrimSpace(line[:idx]))
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", src, err)
	}

	if opts.Sorted {
		sortStrings(results)
	}
	return results, nil
}
