package vault

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PlaceholderOptions controls how placeholders are generated.
type PlaceholderOptions struct {
	// Src is the source .env file to read keys from.
	Src string
	// Dst is the output path for the placeholder file.
	// Defaults to <Src>.placeholder if empty.
	Dst string
	// Overwrite allows overwriting an existing destination file.
	Overwrite bool
	// ValueFmt is the format string used for placeholder values.
	// It receives the key name as its argument. Defaults to "<KEY>".
	ValueFmt string
}

// GeneratePlaceholder reads a .env file and writes a new file where every
// value is replaced by a descriptive placeholder, making it safe to commit
// as documentation without leaking secrets.
func GeneratePlaceholder(cfg Config, opts PlaceholderOptions) (string, error) {
	if !IsInitialised(cfg) {
		return "", fmt.Errorf("vault is not initialised; run 'envault init' first")
	}

	src := opts.Src
	if src == "" {
		src = ".env"
	}

	dst := opts.Dst
	if dst == "" {
		ext := filepath.Ext(src)
		base := strings.TrimSuffix(src, ext)
		dst = base + ".placeholder" + ext
	}

	if !opts.Overwrite {
		if _, err := os.Stat(dst); err == nil {
			return "", fmt.Errorf("destination already exists: %s (use --overwrite to replace)", dst)
		}
	}

	f, err := os.Open(src)
	if err != nil {
		return "", fmt.Errorf("open source: %w", err)
	}
	defer f.Close()

	valueFmt := opts.ValueFmt
	if valueFmt == "" {
		valueFmt = "<%s>"
	}

	var out strings.Builder
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			out.WriteString(line + "\n")
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			out.WriteString(line + "\n")
			continue
		}
		key := strings.TrimSpace(parts[0])
		placeholder := fmt.Sprintf(valueFmt, key)
		out.WriteString(fmt.Sprintf("%s=%s\n", key, placeholder))
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scan source: %w", err)
	}

	if err := os.WriteFile(dst, []byte(out.String()), 0o644); err != nil {
		return "", fmt.Errorf("write placeholder file: %w", err)
	}
	return dst, nil
}
