package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// CountResult holds the key/value statistics for an env file.
type CountResult struct {
	Total    int
	Set      int
	Empty    int
	Comments int
	Blanks   int
}

// Count reads the env file at src and returns a CountResult with
// statistics about the number of keys, empty values, comments, and
// blank lines. Returns an error if the vault is not initialised or
// the file cannot be read.
func Count(cfg Config, src string) (CountResult, error) {
	if !IsInitialised(cfg) {
		return CountResult{}, fmt.Errorf("vault not initialised: run 'envault init' first")
	}

	if src == "" {
		src = ".env"
	}

	f, err := os.Open(src)
	if err != nil {
		return CountResult{}, fmt.Errorf("open %s: %w", src, err)
	}
	defer f.Close()

	var result CountResult
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		switch {
		case trimmed == "":
			result.Blanks++
		case strings.HasPrefix(trimmed, "#"):
			result.Comments++
		default:
			result.Total++
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
				result.Set++
			} else {
				result.Empty++
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return CountResult{}, fmt.Errorf("reading %s: %w", src, err)
	}

	return result, nil
}
