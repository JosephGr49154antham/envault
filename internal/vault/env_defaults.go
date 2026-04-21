package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// DefaultsOptions controls behaviour of the Defaults function.
type DefaultsOptions struct {
	// Src is the .env file to fill defaults into (modified in place or written to Dst).
	Src string
	// Defaults is the file containing default key=value pairs.
	Defaults string
	// Dst is the output path. If empty, Src is overwritten.
	Dst string
	// Overwrite replaces existing values in Src with those from Defaults.
	Overwrite bool
}

// Defaults fills missing (or all, when Overwrite is set) keys in Src from a
// defaults file. Keys already present in Src are left untouched unless
// Overwrite is true. Comments and blank lines in Src are preserved.
func Defaults(cfg Config, opts DefaultsOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault is not initialised: run 'envault init' first")
	}

	if opts.Src == "" {
		return fmt.Errorf("src file must be specified")
	}
	if opts.Defaults == "" {
		return fmt.Errorf("defaults file must be specified")
	}

	dst := opts.Dst
	if dst == "" {
		dst = opts.Src
	}

	// Load default values.
	defaultMap, err := loadDefaultsMap(opts.Defaults)
	if err != nil {
		return fmt.Errorf("reading defaults file: %w", err)
	}

	// Read source lines, tracking which keys are already set.
	srcLines, err := readDefaultsSrcLines(opts.Src)
	if err != nil {
		return fmt.Errorf("reading src file: %w", err)
	}

	// Apply defaults: update existing lines if Overwrite, then append missing keys.
	presentKeys := map[string]bool{}
	result := make([]string, 0, len(srcLines)+len(defaultMap))

	for _, line := range srcLines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			result = append(result, line)
			continue
		}
		key, _, ok := parseDefaultsLine(trimmed)
		if !ok {
			result = append(result, line)
			continue
		}
		presentKeys[key] = true
		if opts.Overwrite {
			if val, found := defaultMap[key]; found {
				result = append(result, key+"="+val)
				continue
			}
		}
		result = append(result, line)
	}

	// Append keys from defaults that were not present in src.
	for _, key := range orderedDefaultKeys(opts.Defaults) {
		if !presentKeys[key] {
			result = append(result, key+"="+defaultMap[key])
		}
	}

	// Write output.
	f, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, line := range result {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return w.Flush()
}

// loadDefaultsMap returns a map of key -> value from the given file,
// skipping comments and blank lines.
func loadDefaultsMap(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	m := map[string]string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := parseDefaultsLine(line)
		if ok {
			m[key] = val
		}
	}
	return m, scanner.Err()
}

// orderedDefaultKeys returns keys from a defaults file in file order,
// preserving insertion order for deterministic appending.
func orderedDefaultKeys(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var keys []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if key, _, ok := parseDefaultsLine(line); ok {
			keys = append(keys, key)
		}
	}
	return keys
}

// readDefaultsSrcLines reads all lines from path, returning them verbatim.
func readDefaultsSrcLines(path string) ([]string, error) {
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

// parseDefaultsLine splits a KEY=VALUE line into its parts.
func parseDefaultsLine(line string) (key, value string, ok bool) {
	idx := strings.IndexByte(line, '=')
	if idx < 1 {
		return "", "", false
	}
	return strings.TrimSpace(line[:idx]), line[idx+1:], true
}
