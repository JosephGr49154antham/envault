package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// PatchOptions controls the behaviour of Patch.
type PatchOptions struct {
	// Src is the .env file to patch. Defaults to DefaultConfig().PlainFile.
	Src string
	// Dst is where the patched file is written. Defaults to Src (in-place).
	Dst string
	// Deletions lists keys to remove from the file.
	Deletions []string
	// Upserts maps keys to their new or updated values.
	Upserts map[string]string
}

// Patch applies a set of upserts and deletions to an .env file atomically.
// Lines that are comments or blank are preserved unless a deletion explicitly
// targets the key on the following assignment line.
func Patch(cfg Config, opts PatchOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault not initialised at %s", cfg.VaultDir)
	}

	if opts.Src == "" {
		opts.Src = cfg.PlainFile
	}
	if opts.Dst == "" {
		opts.Dst = opts.Src
	}

	lines, err := readPatchLines(opts.Src)
	if err != nil {
		return fmt.Errorf("patch: read %s: %w", opts.Src, err)
	}

	delSet := make(map[string]bool, len(opts.Deletions))
	for _, k := range opts.Deletions {
		delSet[strings.ToUpper(k)] = true
	}

	upserted := make(map[string]bool, len(opts.Upserts))
	var out []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			out = append(out, line)
			continue
		}
		key := envKey(trimmed)
		uKey := strings.ToUpper(key)
		if delSet[uKey] {
			continue
		}
		if newVal, ok := opts.Upserts[key]; ok {
			out = append(out, key+"="+newVal)
			upserted[key] = true
			continue
		}
		out = append(out, line)
	}

	// Append any upserts that were not already present in the file.
	for k, v := range opts.Upserts {
		if !upserted[k] {
			out = append(out, k+"="+v)
		}
	}

	return writePatchLines(opts.Dst, out)
}

func readPatchLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}

func writePatchLines(path string, lines []string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, l := range lines {
		fmt.Fprintln(w, l)
	}
	return w.Flush()
}

// envKey extracts the key from a KEY=VALUE line.
func envKey(line string) string {
	if idx := strings.IndexByte(line, '='); idx >= 0 {
		return strings.TrimSpace(line[:idx])
	}
	return line
}
