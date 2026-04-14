package vault

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// GrepResult holds a single matched line from an env file.
type GrepResult struct {
	File    string
	LineNum int
	Key     string
	Value   string
	Line    string
}

// GrepOptions controls the behaviour of Grep.
type GrepOptions struct {
	// Pattern is the regex applied against keys (and values if SearchValues is true).
	Pattern string
	// SearchValues also matches against the value side of each assignment.
	SearchValues bool
	// CaseSensitive disables automatic case-folding.
	CaseSensitive bool
}

// Grep searches one or more .env files for keys (and optionally values)
// matching the given regex pattern. Files must reside inside an initialised
// vault directory.
func Grep(cfg Config, opts GrepOptions, files ...string) ([]GrepResult, error) {
	if !IsInitialised(cfg) {
		return nil, fmt.Errorf("vault not initialised: run 'envault init' first")
	}

	pattern := opts.Pattern
	if !opts.CaseSensitive {
		pattern = "(?i)" + pattern
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern %q: %w", opts.Pattern, err)
	}

	var results []GrepResult
	for _, f := range files {
		matches, err := grepFile(f, re, opts.SearchValues)
		if err != nil {
			return nil, err
		}
		results = append(results, matches...)
	}
	return results, nil
}

func grepFile(path string, re *regexp.Regexp, searchValues bool) ([]GrepResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	var results []GrepResult
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		matched := re.MatchString(key)
		if !matched && searchValues {
			matched = re.MatchString(val)
		}
		if matched {
			results = append(results, GrepResult{
				File:    path,
				LineNum: lineNum,
				Key:     key,
				Value:   val,
				Line:    line,
			})
		}
	}
	return results, scanner.Err()
}
