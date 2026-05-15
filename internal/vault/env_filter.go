package vault

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FilterOptions controls how Filter behaves.
type FilterOptions struct {
	Src       string
	Dst       string
	Pattern   string
	Negate    bool // keep lines that do NOT match
	Overwrite bool
}

// Filter reads a .env file and writes only the lines whose keys match
// the given pattern (prefix match). When Negate is true the logic is
// inverted: lines whose keys match are excluded.
func Filter(cfg Config, opts FilterOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault not initialised: run 'envault init' first")
	}

	if opts.Src == "" {
		opts.Src = ".env"
	}
	if opts.Dst == "" {
		exts := filepath.Ext(opts.Src)
		base := strings.TrimSuffix(filepath.Base(opts.Src), exts)
		opts.Dst = filepath.Join(filepath.Dir(opts.Src), base+".filtered"+exts)
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

	var kept []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			kept = append(kept, line)
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		matches := strings.HasPrefix(strings.ToUpper(key), strings.ToUpper(opts.Pattern))
		if matches != opts.Negate {
			kept = append(kept, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan %q: %w", opts.Src, err)
	}

	out, err := os.Create(opts.Dst)
	if err != nil {
		return fmt.Errorf("create %q: %w", opts.Dst, err)
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	for _, l := range kept {
		fmt.Fprintln(w, l)
	}
	return w.Flush()
}
