package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// SubtractOptions controls the behaviour of the Subtract operation.
type SubtractOptions struct {
	// Src is the primary .env file whose keys will be filtered.
	Src string
	// Exclude is the .env file whose keys will be removed from Src.
	Exclude string
	// Dst is the output path. Defaults to <src>.subtracted if empty.
	Dst string
	// Overwrite allows an existing Dst file to be replaced.
	Overwrite bool
}

// Subtract writes a new .env file containing all key=value pairs from Src
// whose keys do NOT appear in Exclude. Comments and blank lines from Src are
// preserved. If Dst is empty, the output path is derived from Src by appending
// ".subtracted".
func Subtract(cfg Config, opts SubtractOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault not initialised; run 'envault init' first")
	}

	if opts.Src == "" {
		return fmt.Errorf("src path is required")
	}
	if opts.Exclude == "" {
		return fmt.Errorf("exclude path is required")
	}

	dst := opts.Dst
	if dst == "" {
		dst = opts.Src + ".subtracted"
	}

	if !opts.Overwrite {
		if _, err := os.Stat(dst); err == nil {
			return fmt.Errorf("destination file already exists: %s (use --overwrite to replace)", dst)
		}
	}

	// Build the set of keys to exclude.
	excludeKeys, err := loadSubtractKeys(opts.Exclude)
	if err != nil {
		return fmt.Errorf("reading exclude file: %w", err)
	}

	// Read Src and filter lines.
	srcFile, err := os.Open(opts.Src)
	if err != nil {
		return fmt.Errorf("opening src file: %w", err)
	}
	defer srcFile.Close()

	var kept []string
	scanner := bufio.NewScanner(srcFile)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Always keep comments and blank lines.
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			kept = append(kept, line)
			continue
		}

		key := subtractKey(trimmed)
		if _, excluded := excludeKeys[key]; excluded {
			continue
		}
		kept = append(kept, line)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanning src file: %w", err)
	}

	return writeSubtractLines(dst, kept)
}

// loadSubtractKeys reads an .env file and returns a set of its keys.
func loadSubtractKeys(path string) (map[string]struct{}, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	keys := make(map[string]struct{})
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if k := subtractKey(line); k != "" {
			keys[k] = struct{}{}
		}
	}
	return keys, scanner.Err()
}

// subtractKey extracts the key portion of a KEY=VALUE line.
func subtractKey(line string) string {
	line = strings.TrimPrefix(line, "export ")
	if idx := strings.IndexByte(line, '='); idx > 0 {
		return strings.TrimSpace(line[:idx])
	}
	return ""
}

// writeSubtractLines writes lines to dst, creating or truncating the file.
func writeSubtractLines(dst string, lines []string) error {
	f, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("creating destination file: %w", err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, l := range lines {
		if _, err := fmt.Fprintln(w, l); err != nil {
			return err
		}
	}
	return w.Flush()
}
