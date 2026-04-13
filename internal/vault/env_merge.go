package vault

import (
	"fmt"
	"os"
	"strings"
)

// MergeResult holds the outcome of merging two env files.
type MergeResult struct {
	Merged  map[string]string // final merged key/value pairs
	Added   []string          // keys present only in src
	Updated []string          // keys in both but with different values
	Kept    []string          // keys unchanged from base
}

// Merge combines a base env file with a source env file.
// Keys in src that are missing from base are added.
// Keys present in both are updated to the src value.
// Keys only in base are kept unchanged.
// Conflicts (differing values) are recorded in Updated.
func Merge(cfg Config, basePath, srcPath, dstPath string) (MergeResult, error) {
	if !IsInitialised(cfg) {
		return MergeResult{}, fmt.Errorf("vault not initialised: run 'envault init' first")
	}

	baseMap, err := parseEnvFile(basePath)
	if err != nil {
		return MergeResult{}, fmt.Errorf("reading base file: %w", err)
	}

	srcMap, err := parseEnvFile(srcPath)
	if err != nil {
		return MergeResult{}, fmt.Errorf("reading source file: %w", err)
	}

	result := MergeResult{
		Merged: make(map[string]string),
	}

	// Start with all base keys.
	for k, v := range baseMap {
		result.Merged[k] = v
	}

	for k, sv := range srcMap {
		bv, exists := baseMap[k]
		switch {
		case !exists:
			result.Added = append(result.Added, k)
			result.Merged[k] = sv
		case bv != sv:
			result.Updated = append(result.Updated, k)
			result.Merged[k] = sv
		default:
			result.Kept = append(result.Kept, k)
		}
	}

	if dstPath == "" {
		dstPath = basePath
	}

	if err := writeMergedEnv(dstPath, result.Merged); err != nil {
		return MergeResult{}, fmt.Errorf("writing merged file: %w", err)
	}

	return result, nil
}

func writeMergedEnv(path string, m map[string]string) error {
	var sb strings.Builder
	for k, v := range m {
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(v)
		sb.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(sb.String()), 0600)
}
