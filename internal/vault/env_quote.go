package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// QuoteOptions controls how values are quoted.
type QuoteOptions struct {
	Src       string
	Dst       string
	Overwrite bool
	Force     bool // quote all values, even those already quoted
}

// Quote rewrites a .env file so that every value is wrapped in double quotes.
// Values that are already quoted are left unchanged unless Force is set.
func Quote(cfg Config, opts QuoteOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault not initialised: run 'envault init' first")
	}

	if opts.Dst == "" {
		opts.Dst = opts.Src
	}

	if !opts.Overwrite && opts.Dst != opts.Src {
		if _, err := os.Stat(opts.Dst); err == nil {
			return fmt.Errorf("destination already exists: %s (use --overwrite to replace)", opts.Dst)
		}
	}

	f, err := os.Open(opts.Src)
	if err != nil {
		return fmt.Errorf("open %s: %w", opts.Src, err)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, quoteLine(scanner.Text(), opts.Force))
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read %s: %w", opts.Src, err)
	}

	out, err := os.Create(opts.Dst)
	if err != nil {
		return fmt.Errorf("create %s: %w", opts.Dst, err)
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	for _, l := range lines {
		fmt.Fprintln(w, l)
	}
	return w.Flush()
}

// quoteLine quotes the value part of a KEY=VALUE line.
func quoteLine(line string, force bool) string {
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

	// Already quoted?
	if !force && len(val) >= 2 &&
		((val[0] == '"' && val[len(val)-1] == '"') ||
			(val[0] == '\'' && val[len(val)-1] == '\'')) {
		return line
	}

	// Strip existing outer quotes when force is set before re-quoting.
	if force && len(val) >= 2 &&
		((val[0] == '"' && val[len(val)-1] == '"') ||
			(val[0] == '\'' && val[len(val)-1] == '\'')) {
		val = val[1 : len(val)-1]
	}

	escaped := strings.ReplaceAll(val, `"`, `\"`)
	return fmt.Sprintf("%s=\"%s\"", key, escaped)
}
