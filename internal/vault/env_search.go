package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// SearchResult holds a single matching line from an env file.
type SearchResult struct {
	File  string
	Line  int
	Key   string
	Value string
}

// SearchOptions controls how Search behaves.
type SearchOptions struct {
	// Pattern is the substring or key name to search for.
	Pattern string
	// SearchValues includes values in the search when true.
	SearchValues bool
	// CaseSensitive controls case matching.
	CaseSensitive bool
}

// Search scans one or more env files for lines whose key (or optionally value)
// contains the given pattern. Returns matched results or an error.
func Search(cfg Config, opts SearchOptions, files ...string) ([]SearchResult, error) {
	if !IsInitialised(cfg) {
		return nil, fmt.Errorf("vault not initialised; run 'envault init' first")
	}
	if opts.Pattern == "" {
		return nil, fmt.Errorf("search pattern must not be empty")
	}

	pattern := opts.Pattern
	if !opts.CaseSensitive {
		pattern = strings.ToLower(pattern)
	}

	var results []SearchResult
	for _, file := range files {
		matches, err := searchFile(file, pattern, opts)
		if err != nil {
			return nil, fmt.Errorf("searching %s: %w", file, err)
		}
		results = append(results, matches...)
	}
	return results, nil
}

func searchFile(path, pattern string, opts SearchOptions) ([]SearchResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var results []SearchResult
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
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(trimmed[:idx])
		value := strings.TrimSpace(trimmed[idx+1:])

		haystack := key
		if opts.SearchValues {
			haystack = key + "=" + value
		}
		if !opts.CaseSensitive {
			haystack = strings.ToLower(haystack)
		}
		if strings.Contains(haystack, pattern) {
			results = append(results, SearchResult{
				File:  path,
				Line:  lineNum,
				Key:   key,
				Value: value,
			})
		}
	}
	return results, scanner.Err()
}
