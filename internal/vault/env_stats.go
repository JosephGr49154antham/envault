package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// StatsResult holds aggregate statistics about a .env file.
type StatsResult struct {
	TotalLines    int
	KeyValuePairs int
	CommentLines  int
	BlankLines    int
	EmptyValues   int
	UniqueKeys    int
	DuplicateKeys int
}

// Stats analyses a .env file and returns aggregate statistics.
func Stats(cfg Config, src string) (*StatsResult, error) {
	if !IsInitialised(cfg) {
		return nil, fmt.Errorf("vault not initialised; run 'envault init' first")
	}

	if src == "" {
		src = ".env"
	}

	f, err := os.Open(src)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", src, err)
	}
	defer f.Close()

	result := &StatsResult{}
	seen := make(map[string]int)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		result.TotalLines++

		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == "":
			result.BlankLines++
		case strings.HasPrefix(trimmed, "#"):
			result.CommentLines++
		default:
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 {
				result.KeyValuePairs++
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				if val == "" {
					result.EmptyValues++
				}
				seen[key]++
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading %s: %w", src, err)
	}

	for _, count := range seen {
		result.UniqueKeys++
		if count > 1 {
			result.DuplicateKeys++
		}
	}

	return result, nil
}
