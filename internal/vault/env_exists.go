package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExistsResult holds the result of checking whether a key exists in an env file.
type ExistsResult struct {
	Key   string
	Found bool
	Value string
}

// ExistsOptions controls the behaviour of the Exists function.
type ExistsOptions struct {
	// ShowValue includes the value in the result when true.
	ShowValue bool
}

// Exists checks whether one or more keys are present in the given env file.
// It returns a result for each queried key.
func Exists(cfg Config, src string, keys []string, opts ExistsOptions) ([]ExistsResult, error) {
	if !IsInitialised(cfg) {
		return nil, fmt.Errorf("vault is not initialised; run 'envault init' first")
	}

	if src == "" {
		src = ".env"
	}

	if !filepath.IsAbs(src) {
		src = filepath.Join(cfg.VaultDir, src)
	}

	data, err := os.ReadFile(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("env file not found: %s", src)
		}
		return nil, fmt.Errorf("read env file: %w", err)
	}

	envMap := parseExistsLines(strings.Split(string(data), "\n"))

	results := make([]ExistsResult, 0, len(keys))
	for _, key := range keys {
		val, found := envMap[key]
		r := ExistsResult{Key: key, Found: found}
		if found && opts.ShowValue {
			r.Value = val
		}
		results = append(results, r)
	}
	return results, nil
}

// parseExistsLines builds a key→value map from raw env file lines,
// skipping blank lines and comments.
func parseExistsLines(lines []string) map[string]string {
	m := make(map[string]string, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		idx := strings.IndexByte(trimmed, '=')
		if idx < 1 {
			continue
		}
		key := strings.TrimSpace(trimmed[:idx])
		val := strings.TrimSpace(trimmed[idx+1:])
		val = strings.Trim(val, `"`)
		m[key] = val
	}
	return m
}
