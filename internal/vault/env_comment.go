package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// CommentOptions configures the Comment operation.
type CommentOptions struct {
	Src     string
	Dst     string
	Keys    []string
	Uncomment bool
	Overwrite bool
}

// Comment adds or removes inline comments on matching key lines in a .env file.
func Comment(cfg Config, opts CommentOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault not initialised: run 'envault init' first")
	}

	if opts.Dst == "" {
		if opts.Overwrite {
			opts.Dst = opts.Src
		} else {
			opts.Dst = strings.TrimSuffix(opts.Src, ".env") + ".commented.env"
		}
	}

	if _, err := os.Stat(opts.Src); os.IsNotExist(err) {
		return fmt.Errorf("source file not found: %s", opts.Src)
	}

	if !opts.Overwrite {
		if _, err := os.Stat(opts.Dst); err == nil {
			return fmt.Errorf("destination already exists: %s (use overwrite flag)", opts.Dst)
		}
	}

	keySet := make(map[string]bool, len(opts.Keys))
	for _, k := range opts.Keys {
		keySet[strings.TrimSpace(k)] = true
	}

	f, err := os.Open(opts.Src)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer f.Close()

	var out []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		out = append(out, applyComment(line, keySet, opts.Uncomment))
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read source: %w", err)
	}

	return writeCommentLines(opts.Dst, out)
}

func applyComment(line string, keySet map[string]bool, uncomment bool) string {
	trimmed := strings.TrimSpace(line)
	if uncomment {
		if !strings.HasPrefix(trimmed, "#") {
			return line
		}
		candidate := strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
		if idx := strings.Index(candidate, "="); idx > 0 {
			key := strings.TrimSpace(candidate[:idx])
			if len(keySet) == 0 || keySet[key] {
				return candidate
			}
		}
		return line
	}
	// comment out
	if strings.HasPrefix(trimmed, "#") || trimmed == "" {
		return line
	}
	if idx := strings.Index(trimmed, "="); idx > 0 {
		key := strings.TrimSpace(trimmed[:idx])
		if len(keySet) == 0 || keySet[key] {
			return "# " + trimmed
		}
	}
	return line
}

func writeCommentLines(dst string, lines []string) error {
	f, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create destination: %w", err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, l := range lines {
		fmt.Fprintln(w, l)
	}
	return w.Flush()
}
