package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Token represents a single parsed element from a .env file.
type Token struct {
	Line    int
	Kind    string // "key", "comment", "blank", "invalid"
	Key     string
	Value   string
	Raw     string
}

// TokenizeResult holds the output of a Tokenize call.
type TokenizeResult struct {
	Tokens []Token
	Errors []string
}

// Tokenize parses a .env file into a slice of Tokens, categorising
// each line as a key=value assignment, comment, blank, or invalid.
func Tokenize(cfg Config, src string) (*TokenizeResult, error) {
	if !IsInitialised(cfg) {
		return nil, fmt.Errorf("vault is not initialised")
	}

	f, err := os.Open(src)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", src, err)
	}
	defer f.Close()

	result := &TokenizeResult{}
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		raw := scanner.Text()
		trimmed := strings.TrimSpace(raw)

		var tok Token
		tok.Line = lineNum
		tok.Raw = raw

		switch {
		case trimmed == "":
			tok.Kind = "blank"
		case strings.HasPrefix(trimmed, "#"):
			tok.Kind = "comment"
		case strings.Contains(trimmed, "="):
			parts := strings.SplitN(trimmed, "=", 2)
			key := strings.TrimSpace(parts[0])
			if key == "" {
				tok.Kind = "invalid"
				result.Errors = append(result.Errors, fmt.Sprintf("line %d: empty key", lineNum))
			} else {
				tok.Kind = "key"
				tok.Key = key
				tok.Value = strings.Trim(strings.TrimSpace(parts[1]), `"`)
			}
		default:
			tok.Kind = "invalid"
			result.Errors = append(result.Errors, fmt.Sprintf("line %d: no '=' found", lineNum))
		}

		result.Tokens = append(result.Tokens, tok)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan %s: %w", src, err)
	}

	return result, nil
}
