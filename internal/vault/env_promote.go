package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PromoteResult holds the outcome of a promotion between two environments.
type PromoteResult struct {
	Src        string
	Dst        string
	KeysAdded  []string
	KeysKept   []string
}

// Promote copies keys from a source env file into a destination env file,
// adding only keys that are absent in the destination. Existing keys in dst
// are never overwritten. Both src and dst are plain-text .env paths.
//
// If dst does not exist it is created. Returns a PromoteResult describing
// which keys were added and which were already present.
func Promote(cfg Config, src, dst string) (PromoteResult, error) {
	if !IsInitialised(cfg) {
		return PromoteResult{}, fmt.Errorf("vault is not initialised")
	}

	srcMap, err := readEnvFileToMap(src)
	if err != nil {
		return PromoteResult{}, fmt.Errorf("reading src %s: %w", src, err)
	}

	dstMap := map[string]string{}
	if _, statErr := os.Stat(dst); statErr == nil {
		dstMap, err = readEnvFileToMap(dst)
		if err != nil {
			return PromoteResult{}, fmt.Errorf("reading dst %s: %w", dst, err)
		}
	}

	result := PromoteResult{Src: src, Dst: dst}
	merged := make(map[string]string, len(dstMap))
	for k, v := range dstMap {
		merged[k] = v
	}

	for k, v := range srcMap {
		if _, exists := dstMap[k]; exists {
			result.KeysKept = append(result.KeysKept, k)
		} else {
			merged[k] = v
			result.KeysAdded = append(result.KeysAdded, k)
		}
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return PromoteResult{}, fmt.Errorf("creating dst dir: %w", err)
	}

	if err := writePromotedEnv(dst, merged); err != nil {
		return PromoteResult{}, fmt.Errorf("writing dst: %w", err)
	}

	return result, nil
}

func writePromotedEnv(path string, env map[string]string) error {
	var sb strings.Builder
	for k, v := range env {
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(v)
		sb.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(sb.String()), 0o600)
}
