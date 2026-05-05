package vault

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// NormalizeOptions controls how normalization is applied.
type NormalizeOptions struct {
	// TrimValues removes leading/trailing whitespace from values.
	TrimValues bool
	// UpperKeys converts all keys to UPPER_CASE.
	UpperKeys bool
	// RemoveBlanks strips blank lines from the output.
	RemoveBlanks bool
	// RemoveComments strips comment lines from the output.
	RemoveComments bool
	// QuoteValues wraps unquoted values that contain spaces in double quotes.
	QuoteValues bool
	// Dst is the output path; if empty the source file is overwritten.
	Dst string
	// Overwrite allows overwriting an existing destination file.
	Overwrite bool
}

// Normalize applies a configurable set of normalisation passes to a .env file
// and writes the result to Dst (or overwrites Src when Dst is empty).
func Normalize(cfg Config, src string, opts NormalizeOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault is not initialised; run 'envault init' first")
	}

	if src == "" {
		src = ".env"
	}

	dst := opts.Dst
	if dst == "" {
		dst = src
	}

	if dst != src && !opts.Overwrite {
		if _, err := os.Stat(dst); err == nil {
			return fmt.Errorf("destination file %q already exists; use --overwrite to replace it", dst)
		}
	}

	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %q: %w", src, err)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read %q: %w", src, err)
	}

	normalized := normalizeLines(lines, opts)

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create %q: %w", dst, err)
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	for _, line := range normalized {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

// normalizeLines applies the requested transformations to a slice of raw lines.
func normalizeLines(lines []string, opts NormalizeOptions) []string {
	var out []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Blank line handling.
		if trimmed == "" {
			if !opts.RemoveBlanks {
				out = append(out, "")
			}
			continue
		}

		// Comment line handling.
		if strings.HasPrefix(trimmed, "#") {
			if !opts.RemoveComments {
				out = append(out, line)
			}
			continue
		}

		// Key=Value line.
		eqIdx := strings.IndexByte(trimmed, '=')
		if eqIdx < 0 {
			// Not a valid KV line — pass through unchanged.
			out = append(out, line)
			continue
		}

		key := trimmed[:eqIdx]
		val := trimmed[eqIdx+1:]

		if opts.UpperKeys {
			key = strings.ToUpper(key)
		}

		if opts.TrimValues {
			val = strings.TrimSpace(val)
		}

		if opts.QuoteValues {
			val = maybeQuoteNormalize(val)
		}

		out = append(out, key+"="+val)
	}
	return out
}

// maybeQuoteNormalize wraps val in double quotes if it contains whitespace and
// is not already quoted.
func maybeQuoteNormalize(val string) string {
	if len(val) >= 2 {
		if (val[0] == '"' && val[len(val)-1] == '"') ||
			(val[0] == '\'' && val[len(val)-1] == '\'') {
			return val
		}
	}
	if strings.ContainsAny(val, " \t") {
		return `"` + val + `"`
	}
	return val
}
