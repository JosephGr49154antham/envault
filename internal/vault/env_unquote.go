package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Unquote reads an env file and strips surrounding quotes from values,
// writing the result to dst. If dst is empty, the source file is overwritten.
func Unquote(cfg Config, src, dst string, overwrite bool) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault is not initialised")
	}

	if src == "" {
		src = ".env"
	}
	if dst == "" {
		dst = src
	}

	if !overwrite && dst != src {
		if _, err := os.Stat(dst); err == nil {
			return fmt.Errorf("destination %q already exists; use --overwrite to replace", dst)
		}
	}

	lines, err := readUnquoteLines(src)
	if err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create %q: %w", dst, err)
	}
	defer out.Close()

	for _, line := range lines {
		fmt.Fprintln(out, line)
	}
	return nil
}

func readUnquoteLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}
	defer f.Close()

	var out []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		out = append(out, unquoteLine(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// unquoteLine strips surrounding single or double quotes from the value part
// of a KEY=VALUE line. Comments and blank lines are passed through unchanged.
func unquoteLine(line string) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return line
	}
	idx := strings.IndexByte(trimmed, '=')
	if idx < 0 {
		return line
	}
	key := trimmed[:idx]
	val := trimmed[idx+1:]

	if len(val) >= 2 {
		if (val[0] == '"' && val[len(val)-1] == '"') ||
			(val[0] == '\'' && val[len(val)-1] == '\'') {
			val = val[1 : len(val)-1]
		}
	}
	return key + "=" + val
}
