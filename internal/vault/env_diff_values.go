package vault

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

// DiffValuesResult holds the outcome of comparing values between two env files.
type DiffValuesResult struct {
	// Changed contains keys whose values differ between the two files.
	Changed map[string][2]string // key -> [leftValue, rightValue]
	// OnlyInLeft contains keys present only in the left file.
	OnlyInLeft []string
	// OnlyInRight contains keys present only in the right file.
	OnlyInRight []string
}

// DiffValues compares the values of two .env files and returns a structured
// result describing which keys have changed values, and which keys exist in
// only one of the files. Comments and blank lines are ignored.
func DiffValues(cfg Config, leftPath, rightPath string) (*DiffValuesResult, error) {
	if !IsInitialised(cfg) {
		return nil, fmt.Errorf("vault is not initialised; run 'envault init' first")
	}

	leftMap, err := readEnvMapFromFile(leftPath)
	if err != nil {
		return nil, fmt.Errorf("reading left file %q: %w", leftPath, err)
	}

	rightMap, err := readEnvMapFromFile(rightPath)
	if err != nil {
		return nil, fmt.Errorf("reading right file %q: %w", rightPath, err)
	}

	result := &DiffValuesResult{
		Changed: make(map[string][2]string),
	}

	for key, leftVal := range leftMap {
		if rightVal, ok := rightMap[key]; ok {
			if leftVal != rightVal {
				result.Changed[key] = [2]string{leftVal, rightVal}
			}
		} else {
			result.OnlyInLeft = append(result.OnlyInLeft, key)
		}
	}

	for key := range rightMap {
		if _, ok := leftMap[key]; !ok {
			result.OnlyInRight = append(result.OnlyInRight, key)
		}
	}

	sort.Strings(result.OnlyInLeft)
	sort.Strings(result.OnlyInRight)

	return result, nil
}

// readEnvMapFromFile parses a .env file into a key→value map.
// Lines starting with '#' and blank lines are skipped.
func readEnvMapFromFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	m := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		// Strip surrounding quotes if present.
		if len(val) >= 2 &&
			((val[0] == '"' && val[len(val)-1] == '"') ||
				(val[0] == '\'' && val[len(val)-1] == '\'')) {
			val = val[1 : len(val)-1]
		}
		m[key] = val
	}
	return m, scanner.Err()
}
