package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// MaskOptions controls Mask behaviour.
type MaskOptions struct {
	// Src is the input .env file path.
	Src string
	// Dst is the output path; defaults to <Src>.masked.
	Dst string
	// Keys is an explicit list of keys to mask. If empty, isSensitive heuristic is used.
	Keys []string
	// MaskChar is the character repeated to form the mask. Defaults to "*".
	MaskChar string
	// Overwrite allows replacing an existing Dst file.
	Overwrite bool
}

// Mask reads Src, replaces sensitive (or explicitly listed) values with a
// redaction mask, and writes the result to Dst.
func Mask(cfg Config, opts MaskOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault is not initialised; run 'envault init' first")
	}

	if opts.Src == "" {
		opts.Src = ".env"
	}
	if opts.Dst == "" {
		opts.Dst = opts.Src + ".masked"
	}
	if opts.MaskChar == "" {
		opts.MaskChar = "*"
	}

	if !opts.Overwrite {
		if _, err := os.Stat(opts.Dst); err == nil {
			return fmt.Errorf("destination %q already exists; use --overwrite to replace", opts.Dst)
		}
	}

	f, err := os.Open(opts.Src)
	if err != nil {
		return fmt.Errorf("open %q: %w", opts.Src, err)
	}
	defer f.Close()

	keySet := make(map[string]bool, len(opts.Keys))
	for _, k := range opts.Keys {
		keySet[strings.ToUpper(k)] = true
	}

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, maskLine(line, keySet, opts.MaskChar))
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read %q: %w", opts.Src, err)
	}

	out, err := os.Create(opts.Dst)
	if err != nil {
		return fmt.Errorf("create %q: %w", opts.Dst, err)
	}
	defer out.Close()

	for _, l := range lines {
		fmt.Fprintln(out, l)
	}
	return nil
}

// maskLine masks the value portion of a KEY=VALUE line when appropriate.
func maskLine(line, keySet map[string]bool, maskChar string) func(string) string {
	return func(line string) string {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			return line
		}
		idx := strings.IndexByte(trimmed, '=')
		if idx < 0 {
			return line
		}
		key := strings.TrimSpace(trimmed[:idx])
		value := trimmed[idx+1:]
		if len(keySet) > 0 {
			if !keySet[strings.ToUpper(key)] {
				return line
			}
		} else if !isSensitive(key) {
			return line
		}
		if value == "" {
			return line
		}
		return key + "=" + strings.Repeat(maskChar, 8)
	}
}(line)
