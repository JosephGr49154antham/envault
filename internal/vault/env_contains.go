package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ContainsResult holds the result of a key lookup in an env file.
type ContainsResult struct {
	Key   string
	Found bool
	Value string
	Line  int
}

// ContainsOptions controls the behaviour of Contains.
type ContainsOptions struct {
	// ShowValue includes the value in the result when true.
	ShowValue bool
}

// Contains checks whether the given keys exist in the env file at src.
// It returns one ContainsResult per requested key.
func Contains(cfg Config, src string, keys []string, opts ContainsOptions) ([]ContainsResult, error) {
	if !IsInitialised(cfg) {
		return nil, fmt.Errorf("vault not initialised: run 'envault init' first")
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("at least one key must be specified")
	}

	f, err := os.Open(src)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", src, err)
	}
	defer f.Close()

	// Build a map from the file for O(1) lookup.
	type entry struct {
		value string
		line  int
	}
	found := make(map[string]entry)

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		raw := scanner.Text()
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		idx := strings.IndexByte(trimmed, '=')
		if idx < 1 {
			continue
		}
		k := strings.TrimSpace(trimmed[:idx])
		v := strings.TrimSpace(trimmed[idx+1:])
		v = strings.Trim(v, `"`)
		found[k] = entry{value: v, line: lineNum}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", src, err)
	}

	results := make([]ContainsResult, 0, len(keys))
	for _, k := range keys {
		e, ok := found[k]
		r := ContainsResult{Key: k, Found: ok, Line: e.line}
		if ok && opts.ShowValue {
			r.Value = e.value
		}
		results = append(results, r)
	}
	return results, nil
}
