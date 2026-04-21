package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// PickOptions controls the behaviour of Pick.
type PickOptions struct {
	// Keys is the list of keys to extract from the source file.
	Keys []string
	// Src is the source .env file path.
	Src string
	// Dst is the destination file path. If empty, defaults to <src>.picked.env.
	Dst string
	// Overwrite allows an existing destination file to be replaced.
	Overwrite bool
}

// Pick extracts a subset of keys from a .env file and writes them to Dst.
// Keys that are not found in Src are silently skipped.
func Pick(cfg Config, opts PickOptions) (string, error) {
	if !IsInitialised(cfg) {
		return "", fmt.Errorf("vault not initialised: run 'envault init' first")
	}

	if opts.Dst == "" {
		base := strings.TrimSuffix(opts.Src, ".env")
		opts.Dst = base + ".picked.env"
	}

	if !opts.Overwrite {
		if _, err := os.Stat(opts.Dst); err == nil {
			return "", fmt.Errorf("destination already exists: %s (use --overwrite to replace)", opts.Dst)
		}
	}

	wantSet := make(map[string]struct{}, len(opts.Keys))
	for _, k := range opts.Keys {
		wantSet[strings.TrimSpace(k)] = struct{}{}
	}

	f, err := os.Open(opts.Src)
	if err != nil {
		return "", fmt.Errorf("open source: %w", err)
	}
	defer f.Close()

	var picked []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		if _, ok := wantSet[key]; ok {
			picked = append(picked, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scan source: %w", err)
	}

	out, err := os.Create(opts.Dst)
	if err != nil {
		return "", fmt.Errorf("create destination: %w", err)
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	for _, line := range picked {
		fmt.Fprintln(w, line)
	}
	if err := w.Flush(); err != nil {
		return "", fmt.Errorf("write destination: %w", err)
	}

	return opts.Dst, nil
}
