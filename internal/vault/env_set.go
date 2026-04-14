package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Set adds or updates a key-value pair in a .env file.
// If the key already exists, its value is updated in place.
// If the key does not exist, it is appended to the end of the file.
// If dst is empty, src is modified in place.
func Set(cfg Config, src, dst, key, value string) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault is not initialised: run 'envault init' first")
	}

	if key == "" {
		return fmt.Errorf("key must not be empty")
	}

	if strings.ContainsAny(key, " \t=") {
		return fmt.Errorf("key %q contains invalid characters", key)
	}

	if dst == "" {
		dst = src
	}

	lines, err := readLines(src)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", src, err)
	}

	updated := false
	prefix := key + "="

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") || trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, prefix) {
			lines[i] = key + "=" + value
			updated = true
			break
		}
	}

	if !updated {
		lines = append(lines, key+"="+value)
	}

	return writeLines(dst, lines)
}

// Unset removes a key from a .env file.
// If the key is not found, no error is returned.
// If dst is empty, src is modified in place.
func Unset(cfg Config, src, dst, key string) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault is not initialised: run 'envault init' first")
	}

	if key == "" {
		return fmt.Errorf("key must not be empty")
	}

	if dst == "" {
		dst = src
	}

	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer f.Close()

	prefix := key + "="
	var kept []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "#") && strings.HasPrefix(trimmed, prefix) {
			continue
		}
		kept = append(kept, line)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan %s: %w", src, err)
	}
	f.Close()

	return writeLines(dst, kept)
}
