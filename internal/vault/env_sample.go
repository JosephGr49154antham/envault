package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// SampleOptions configures the Sample operation.
type SampleOptions struct {
	Src string
	Dst string
	N    int  // number of entries to sample; 0 means all
	Keys bool // if true, output keys only (no values)
}

// Sample reads up to N key=value pairs from Src (skipping comments and blanks)
// and writes them to Dst. If Keys is true, only the key names are written.
func Sample(cfg Config, opts SampleOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault is not initialised; run 'envault init' first")
	}

	if opts.Src == "" {
		opts.Src = cfg.PlainFile
	}
	if opts.Dst == "" {
		opts.Dst = opts.Src + ".sample"
	}

	f, err := os.Open(opts.Src)
	if err != nil {
		return fmt.Errorf("sample: open %s: %w", opts.Src, err)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		eqIdx := strings.IndexByte(trimmed, '=')
		if eqIdx < 0 {
			continue
		}
		if opts.Keys {
			lines = append(lines, trimmed[:eqIdx])
		} else {
			lines = append(lines, trimmed[:eqIdx+1])
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("sample: scan %s: %w", opts.Src, err)
	}

	if opts.N > 0 && opts.N < len(lines) {
		lines = lines[:opts.N]
	}

	out, err := os.Create(opts.Dst)
	if err != nil {
		return fmt.Errorf("sample: create %s: %w", opts.Dst, err)
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	for _, l := range lines {
		fmt.Fprintln(w, l)
	}
	return w.Flush()
}
