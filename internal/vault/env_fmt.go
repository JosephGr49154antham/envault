package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// FmtOptions controls formatting behaviour.
type FmtOptions struct {
	// Dst is the output path; if empty, the source file is overwritten.
	Dst string
	// TrimValues removes surrounding whitespace from values.
	TrimValues bool
	// QuoteValues wraps unquoted values containing spaces in double-quotes.
	QuoteValues bool
}

// Fmt normalises a .env file by cleaning up key=value formatting.
// It preserves comments and blank lines.
func Fmt(cfg Config, src string, opts FmtOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault not initialised: run 'envault init' first")
	}

	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("source file not found: %w", err)
	}

	lines, err := readFmtLines(src)
	if err != nil {
		return err
	}

	formatted := make([]string, 0, len(lines))
	for _, line := range lines {
		formatted = append(formatted, formatLine(line, opts))
	}

	dst := opts.Dst
	if dst == "" {
		dst = src
	}

	f, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("cannot write output: %w", err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, l := range formatted {
		fmt.Fprintln(w, l)
	}
	return w.Flush()
}

func readFmtLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}

func formatLine(line string, opts FmtOptions) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return line
	}

	idx := strings.IndexByte(trimmed, '=')
	if idx < 0 {
		return trimmed
	}

	key := strings.TrimSpace(trimmed[:idx])
	val := trimmed[idx+1:]

	if opts.TrimValues {
		val = strings.TrimSpace(val)
	}

	if opts.QuoteValues {
		unquoted := strings.Trim(val, `"`)
		if strings.ContainsAny(unquoted, " \t") && !strings.HasPrefix(val, `"`) {
			val = `"` + unquoted + `"`
		}
	}

	return key + "=" + val
}
