package vault

import (
	"fmt"
	"os"
	"strings"
)

// CopyOptions controls the behaviour of the Copy operation.
type CopyOptions struct {
	// Keys is the list of keys to copy from Src to Dst.
	// If empty, all keys are copied.
	Keys []string
	// Overwrite allows existing keys in Dst to be replaced.
	Overwrite bool
	// Dst is the destination .env file path.
	// Defaults to ".env" in the vault root.
	Dst string
}

// Copy copies selected (or all) keys from one .env file into another,
// optionally overwriting existing values.
func Copy(cfg Config, src string, opts CopyOptions) (int, error) {
	if !IsInitialised(cfg) {
		return 0, fmt.Errorf("vault is not initialised; run 'envault init' first")
	}

	if opts.Dst == "" {
		opts.Dst = ".env"
	}

	srcMap, err := readEnvFileToMap(src)
	if err != nil {
		return 0, fmt.Errorf("reading source file: %w", err)
	}

	// Build the subset to copy.
	toCopy := srcMap
	if len(opts.Keys) > 0 {
		toCopy = make(map[string]string, len(opts.Keys))
		for _, k := range opts.Keys {
			v, ok := srcMap[k]
			if !ok {
				return 0, fmt.Errorf("key %q not found in %s", k, src)
			}
			toCopy[k] = v
		}
	}

	// Load existing destination (may not exist yet).
	dstMap := map[string]string{}
	if _, err := os.Stat(opts.Dst); err == nil {
		dstMap, err = readEnvFileToMap(opts.Dst)
		if err != nil {
			return 0, fmt.Errorf("reading destination file: %w", err)
		}
	}

	copied := 0
	for k, v := range toCopy {
		if _, exists := dstMap[k]; exists && !opts.Overwrite {
			continue
		}
		dstMap[k] = v
		copied++
	}

	if err := writeCopyEnv(opts.Dst, dstMap); err != nil {
		return 0, fmt.Errorf("writing destination file: %w", err)
	}
	return copied, nil
}

func writeCopyEnv(path string, m map[string]string) error {
	var sb strings.Builder
	for k, v := range m {
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(v)
		sb.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(sb.String()), 0o600)
}
