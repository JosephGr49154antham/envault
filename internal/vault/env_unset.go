package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// UnsetOptions configures the Unset operation.
type UnsetOptions struct {
	Src  string
	Dst  string
	Keys []string
}

// UnsetResult holds the outcome of an Unset operation.
type UnsetResult struct {
	Removed []string
	Missing []string
}

// Unset removes one or more keys from an env file.
// If Dst is empty the source file is updated in place.
func Unset(cfg Config, opts UnsetOptions) (UnsetResult, error) {
	if !IsInitialised(cfg) {
		return UnsetResult{}, fmt.Errorf("vault not initialised")
	}
	if len(opts.Keys) == 0 {
		return UnsetResult{}, fmt.Errorf("no keys specified")
	}

	dst := opts.Dst
	if dst == "" {
		dst = opts.Src
	}

	f, err := os.Open(opts.Src)
	if err != nil {
		return UnsetResult{}, fmt.Errorf("open %s: %w", opts.Src, err)
	}
	defer f.Close()

	targetSet := make(map[string]bool, len(opts.Keys))
	for _, k := range opts.Keys {
		targetSet[strings.TrimSpace(k)] = true
	}

	var kept []string
	found := make(map[string]bool)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			kept = append(kept, line)
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		key := strings.TrimSpace(parts[0])
		if targetSet[key] {
			found[key] = true
			continue // drop the line
		}
		kept = append(kept, line)
	}
	if err := scanner.Err(); err != nil {
		return UnsetResult{}, fmt.Errorf("scan %s: %w", opts.Src, err)
	}

	out, err := os.Create(dst)
	if err != nil {
		return UnsetResult{}, fmt.Errorf("create %s: %w", dst, err)
	}
	defer out.Close()

	for _, line := range kept {
		fmt.Fprintln(out, line)
	}

	var result UnsetResult
	for _, k := range opts.Keys {
		if found[k] {
			result.Removed = append(result.Removed, k)
		} else {
			result.Missing = append(result.Missing, k)
		}
	}
	return result, nil
}
