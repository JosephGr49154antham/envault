package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// GetOptions controls the behaviour of the Get operation.
type GetOptions struct {
	// Src is the path to the .env file to read from.
	// Defaults to the vault's plain .env file.
	Src string

	// Key is the environment variable name to look up.
	Key string

	// Quiet suppresses the key= prefix and prints only the raw value.
	Quiet bool
}

// GetResult holds the outcome of a Get call.
type GetResult struct {
	Key   string
	Value string
	Found bool
}

// Get retrieves the value of a single key from an env file.
// If the key is not present, Found is false and no error is returned.
// An error is returned only for I/O or initialisation failures.
func Get(cfg Config, opts GetOptions) (GetResult, error) {
	if !IsInitialised(cfg) {
		return GetResult{}, fmt.Errorf("vault is not initialised: run 'envault init' first")
	}

	src := opts.Src
	if src == "" {
		src = cfg.PlainFile
	}

	if opts.Key == "" {
		return GetResult{}, fmt.Errorf("key must not be empty")
	}

	f, err := os.Open(src)
	if err != nil {
		if os.IsNotExist(err) {
			return GetResult{Key: opts.Key, Found: false}, nil
		}
		return GetResult{}, fmt.Errorf("open %s: %w", src, err)
	}
	defer f.Close()

	target := strings.ToUpper(strings.TrimSpace(opts.Key))

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		// skip blank lines and comments
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// strip optional "export " prefix
		trimmed = strings.TrimPrefix(trimmed, "export ")

		eqIdx := strings.IndexByte(trimmed, '=')
		if eqIdx < 0 {
			continue
		}

		k := strings.TrimSpace(trimmed[:eqIdx])
		v := strings.TrimSpace(trimmed[eqIdx+1:])

		// strip surrounding quotes if present
		if len(v) >= 2 {
			if (v[0] == '"' && v[len(v)-1] == '"') ||
				(v[0] == '\'' && v[len(v)-1] == '\'') {
				v = v[1 : len(v)-1]
			}
		}

		if strings.ToUpper(k) == target {
			return GetResult{Key: k, Value: v, Found: true}, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return GetResult{}, fmt.Errorf("read %s: %w", src, err)
	}

	return GetResult{Key: opts.Key, Found: false}, nil
}
