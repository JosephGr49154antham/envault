package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// WhitelistOptions configures the Whitelist operation.
type WhitelistOptions struct {
	Src      string
	Dst      string
	Keys     []string
	Overwrite bool
}

// Whitelist reads src and writes only the key=value pairs whose keys appear in
// the provided allow-list to dst. Comments and blank lines are preserved when
// they appear before a whitelisted key.
func Whitelist(cfg Config, opts WhitelistOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault is not initialised")
	}

	if opts.Dst == "" {
		opts.Dst = opts.Src + ".whitelist"
	}

	if !opts.Overwrite {
		if _, err := os.Stat(opts.Dst); err == nil {
			return fmt.Errorf("destination %q already exists; use overwrite flag to replace", opts.Dst)
		}
	}

	allowed := make(map[string]bool, len(opts.Keys))
	for _, k := range opts.Keys {
		allowed[strings.TrimSpace(k)] = true
	}

	f, err := os.Open(opts.Src)
	if err != nil {
		return fmt.Errorf("open src: %w", err)
	}
	defer f.Close()

	var out []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			out = append(out, line)
			continue
		}
		idx := strings.IndexByte(trimmed, '=')
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(trimmed[:idx])
		if allowed[key] {
			out = append(out, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan src: %w", err)
	}

	return os.WriteFile(opts.Dst, []byte(strings.Join(out, "\n")+"\n"), 0o600)
}
