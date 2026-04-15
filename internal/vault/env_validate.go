package vault

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// ValidationIssue describes a single problem found during env file validation.
type ValidationIssue struct {
	Line    int
	Message string
}

// ValidationResult holds the outcome of validating an env file.
type ValidationResult struct {
	File   string
	Issues []ValidationIssue
	Valid  bool
}

var validKeyRe = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)

// Validate checks an env file for common problems:
//   - keys that are not upper-case identifiers
//   - lines that are neither blank, comments, nor KEY=VALUE pairs
//   - empty values (reported as warnings, not errors)
func Validate(cfg Config, src string) (ValidationResult, error) {
	if !IsInitialised(cfg) {
		return ValidationResult{}, fmt.Errorf("vault is not initialised")
	}

	if src == "" {
		src = cfg.PlainFile
	}

	f, err := os.Open(src)
	if err != nil {
		return ValidationResult{}, fmt.Errorf("open %s: %w", src, err)
	}
	defer f.Close()

	result := ValidationResult{File: src, Valid: true}
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		raw := scanner.Text()
		trimmed := strings.TrimSpace(raw)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		eqIdx := strings.IndexByte(trimmed, '=')
		if eqIdx < 0 {
			result.Issues = append(result.Issues, ValidationIssue{
				Line:    lineNum,
				Message: fmt.Sprintf("invalid line (no '=' found): %q", trimmed),
			})
			result.Valid = false
			continue
		}

		key := trimmed[:eqIdx]
		val := trimmed[eqIdx+1:]

		if !validKeyRe.MatchString(key) {
			result.Issues = append(result.Issues, ValidationIssue{
				Line:    lineNum,
				Message: fmt.Sprintf("key %q is not a valid identifier (must be UPPER_CASE)", key),
			})
			result.Valid = false
		}

		if val == "" {
			result.Issues = append(result.Issues, ValidationIssue{
				Line:    lineNum,
				Message: fmt.Sprintf("key %q has an empty value", key),
			})
			// empty values are warnings; do not mark invalid
		}
	}

	if err := scanner.Err(); err != nil {
		return ValidationResult{}, fmt.Errorf("scan %s: %w", src, err)
	}

	return result, nil
}

// Summary returns a human-readable summary of the validation result.
// It reports the number of issues found, or confirms the file is valid.
func (r ValidationResult) Summary() string {
	if r.Valid && len(r.Issues) == 0 {
		return fmt.Sprintf("%s: OK", r.File)
	}
	lines := make([]string, 0, len(r.Issues)+1)
	for _, issue := range r.Issues {
		lines = append(lines, fmt.Sprintf("  line %d: %s", issue.Line, issue.Message))
	}
	status := "warnings"
	if !r.Valid {
		status = "errors"
	}
	return fmt.Sprintf("%s: %d %s\n%s", r.File, len(r.Issues), status, strings.Join(lines, "\n"))
}
