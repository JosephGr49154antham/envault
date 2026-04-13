package vault

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// CompareResult holds the outcome of comparing two env files.
type CompareResult struct {
	OnlyInA    []string          // keys present only in file A
	OnlyInB    []string          // keys present only in file B
	Changed    map[string][2]string // keys present in both but with different values [a, b]
	Identical  []string          // keys with identical values in both files
}

// HasDifferences returns true if there are any differences between the two files.
func (r CompareResult) HasDifferences() bool {
	return len(r.OnlyInA) > 0 || len(r.OnlyInB) > 0 || len(r.Changed) > 0
}

// Compare loads two plain .env files and returns a CompareResult describing
// their differences. Neither file needs to be encrypted.
func Compare(cfg Config, fileA, fileB string) (CompareResult, error) {
	if !IsInitialised(cfg) {
		return CompareResult{}, fmt.Errorf("vault not initialised: run 'envault init' first")
	}

	mapA, err := readEnvFileToMap(fileA)
	if err != nil {
		return CompareResult{}, fmt.Errorf("reading %s: %w", fileA, err)
	}

	mapB, err := readEnvFileToMap(fileB)
	if err != nil {
		return CompareResult{}, fmt.Errorf("reading %s: %w", fileB, err)
	}

	return buildCompareResult(mapA, mapB), nil
}

func readEnvFileToMap(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		m[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return m, nil
}

func buildCompareResult(a, b map[string]string) CompareResult {
	r := CompareResult{
		Changed: make(map[string][2]string),
	}

	for k, va := range a {
		if vb, ok := b[k]; ok {
			if va == vb {
				r.Identical = append(r.Identical, k)
			} else {
				r.Changed[k] = [2]string{va, vb}
			}
		} else {
			r.OnlyInA = append(r.OnlyInA, k)
		}
	}

	for k := range b {
		if _, ok := a[k]; !ok {
			r.OnlyInB = append(r.OnlyInB, k)
		}
	}

	sort.Strings(r.OnlyInA)
	sort.Strings(r.OnlyInB)
	sort.Strings(r.Identical)
	return r
}
