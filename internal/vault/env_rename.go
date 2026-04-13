package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Rename renames an environment variable key within a plain .env file.
// It returns an error if the vault is not initialised, the file is not found,
// or the oldKey does not exist in the file.
func Rename(cfg Config, src, oldKey, newKey string) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault is not initialised: run 'envault init' first")
	}

	if oldKey == "" || newKey == "" {
		return fmt.Errorf("both old and new key names must be non-empty")
	}

	if oldKey == newKey {
		return fmt.Errorf("old and new key names are identical")
	}

	lines, err := readLines(src)
	if err != nil {
		return fmt.Errorf("reading %s: %w", src, err)
	}

	updated, found, dupFound := applyRename(lines, oldKey, newKey)
	if !found {
		return fmt.Errorf("key %q not found in %s", oldKey, src)
	}
	if dupFound {
		return fmt.Errorf("key %q already exists in %s", newKey, src)
	}

	return writeLines(src, updated)
}

func applyRename(lines []string, oldKey, newKey string) (out []string, found, dupFound bool) {
	// pre-scan for newKey collision
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, newKey+"=") {
			dupFound = true
			return lines, false, true
		}
	}

	out = make([]string, len(lines))
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, oldKey+"=") {
			out[i] = newKey + "=" + strings.SplitN(trimmed, "=", 2)[1]
			found = true
		} else {
			out[i] = line
		}
	}
	return out, found, false
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func writeLines(path string, lines []string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}
