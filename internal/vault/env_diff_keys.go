package vault

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// KeyDiffResult holds the result of comparing keys between two env files.
type KeyDiffResult struct {
	OnlyInA []string
	OnlyInB []string
	InBoth  []string
}

// DiffKeys compares the keys (not values) present in two env files and
// returns which keys are exclusive to each file and which are shared.
// Comments and blank lines are ignored.
func DiffKeys(cfg Config, fileA, fileB string) (*KeyDiffResult, error) {
	if !IsInitialised(cfg) {
		return nil, fmt.Errorf("vault not initialised: run 'envault init' first")
	}

	keysA, err := loadEnvKeysFromFile(fileA)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", fileA, err)
	}

	keysB, err := loadEnvKeysFromFile(fileB)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", fileB, err)
	}

	setA := toSet(keysA)
	setB := toSet(keysB)

	result := &KeyDiffResult{}

	for k := range setA {
		if setB[k] {
			result.InBoth = append(result.InBoth, k)
		} else {
			result.OnlyInA = append(result.OnlyInA, k)
		}
	}
	for k := range setB {
		if !setA[k] {
			result.OnlyInB = append(result.OnlyInB, k)
		}
	}

	sort.Strings(result.OnlyInA)
	sort.Strings(result.OnlyInB)
	sort.Strings(result.InBoth)

	return result, nil
}

func loadEnvKeysFromFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var keys []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) >= 1 {
			keys = append(keys, strings.TrimSpace(parts[0]))
		}
	}
	return keys, nil
}

func toSet(keys []string) map[string]bool {
	s := make(map[string]bool, len(keys))
	for _, k := range keys {
		s[k] = true
	}
	return s
}
