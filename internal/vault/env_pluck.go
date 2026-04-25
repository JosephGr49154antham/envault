package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// PluckOptions controls the behaviour of Pluck.
type PluckOptions struct {
	Src  string // source .env file
	Keys []string // keys to extract values for
	Raw  bool // if true, omit the KEY= prefix and print bare values
}

// Pluck reads the given keys from a .env file and writes their values to
// stdout (or returns them as a slice for programmatic use).
// Unlike Pick, Pluck does not write an output file — it is a read-only
// inspection command intended for scripting.
func Pluck(cfg Config, opts PluckOptions) ([]string, error) {
	if !IsInitialised(cfg) {
		return nil, fmt.Errorf("vault not initialised: run 'envault init' first")
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

	want := make(map[string]bool, len(opts.Keys))
	for _, k := range opts.Keys {
		want[strings.TrimSpace(k)] = true
	}

	values := make(map[string]string)
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
		val = strings.Trim(val, `"`)
		if want[key] {
			values[key] = val
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan %s: %w", src, err)
	}

	var results []string
	for _, k := range opts.Keys {
		v, ok := values[k]
		if !ok {
			return nil, fmt.Errorf("key %q not found in %s", k, src)
		}
		if opts.Raw {
			results = append(results, v)
		} else {
			results = append(results, fmt.Sprintf("%s=%s", k, v))
		}
	}
	return results, nil
}
