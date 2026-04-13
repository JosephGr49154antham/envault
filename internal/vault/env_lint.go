package vault

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// LintIssue describes a single linting problem found in an env file.
type LintIssue struct {
	Line    int
	Message string
}

func (i LintIssue) String() string {
	return fmt.Sprintf("line %d: %s", i.Line, i.Message)
}

var validKeyRe = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)

// Lint checks a plain-text .env file for common issues such as:
//   - keys that are not UPPER_SNAKE_CASE
//   - duplicate keys
//   - lines that are neither blank, comments, nor valid KEY=VALUE pairs
//
// It returns a (possibly empty) slice of LintIssue and any I/O error.
func Lint(cfg Config, src string) ([]LintIssue, error) {
	if !IsInitialised(cfg) {
		return nil, fmt.Errorf("vault is not initialised: run 'envault init' first")
	}

	f, err := os.Open(src)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", src, err)
	}
	defer f.Close()

	var issues []LintIssue
	seen := make(map[string]int)

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
		if idx <= 0 {
			issues = append(issues, LintIssue{Line: lineNum, Message: fmt.Sprintf("invalid format %q: expected KEY=VALUE", trimmed)})
			continue
		}

		key := trimmed[:idx]
		if !validKeyRe.MatchString(key) {
			issues = append(issues, LintIssue{Line: lineNum, Message: fmt.Sprintf("key %q should be UPPER_SNAKE_CASE", key)})
		}

		if prev, dup := seen[key]; dup {
			issues = append(issues, LintIssue{Line: lineNum, Message: fmt.Sprintf("duplicate key %q (first seen on line %d)", key, prev)})
		} else {
			seen[key] = lineNum
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading %s: %w", src, err)
	}

	return issues, nil
}
