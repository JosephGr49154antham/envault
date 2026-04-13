package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// RedactResult holds the outcome of a redaction operation.
type RedactResult struct {
	Redacted []string // keys whose values were redacted
	Total    int      // total number of keys processed
}

// Redact reads a plain .env file and writes a copy with secret values
// replaced by "***REDACTED***". Keys matching any pattern in sensitivePatterns
// (case-insensitive substring match) are redacted.
//
// If dst is empty, the default destination is <src>.redacted.
func Redact(cfg Config, src, dst string) (RedactResult, error) {
	if !IsInitialised(cfg) {
		return RedactResult{}, fmt.Errorf("vault not initialised: run 'envault init' first")
	}

	if src == "" {
		src = ".env"
	}
	if dst == "" {
		dst = src + ".redacted"
	}

	in, err := os.Open(src)
	if err != nil {
		return RedactResult{}, fmt.Errorf("open source file: %w", err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return RedactResult{}, fmt.Errorf("create destination file: %w", err)
	}
	defer out.Close()

	var result RedactResult
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := scanner.Text()
		written := redactLine(line, &result)
		fmt.Fprintln(out, written)
	}
	if err := scanner.Err(); err != nil {
		return RedactResult{}, fmt.Errorf("reading source file: %w", err)
	}
	return result, nil
}

// sensitivePatterns lists substrings that mark a key as sensitive.
var sensitivePatterns = []string{
	"secret", "password", "passwd", "token", "apikey", "api_key",
	"private", "credential", "auth", "jwt", "passphrase",
}

func isSensitive(key string) bool {
	lower := strings.ToLower(key)
	for _, p := range sensitivePatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

func redactLine(line string, result *RedactResult) string {
	// Preserve blank lines and comments.
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return line
	}

	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return line
	}

	key := strings.TrimSpace(parts[0])
	result.Total++

	if isSensitive(key) {
		result.Redacted = append(result.Redacted, key)
		return key + "=***REDACTED***"
	}
	return line
}
