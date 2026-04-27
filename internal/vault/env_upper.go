package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// UpperOptions controls the behaviour of the Upper function.
type UpperOptions struct {
	Src  string
	Dst  string
	Keys []string // if empty, all keys are uppercased
}

// Upper normalises env variable keys to UPPER_SNAKE_CASE.
// If Keys is non-empty, only those keys are transformed.
// The result is written to Dst (defaults to Src when empty).
func Upper(cfg Config, opts UpperOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault not initialised: run 'envault init' first")
	}

	lines, err := readUpperLines(opts.Src)
	if err != nil {
		return fmt.Errorf("reading %s: %w", opts.Src, err)
	}

	keySet := make(map[string]bool, len(opts.Keys))
	for _, k := range opts.Keys {
		keySet[strings.ToUpper(k)] = true
	}

	out := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			out = append(out, line)
			continue
		}
		idx := strings.IndexByte(trimmed, '=')
		if idx < 0 {
			out = append(out, line)
			continue
		}
		key := trimmed[:idx]
		val := trimmed[idx+1:]
		upperKey := strings.ToUpper(key)
		if len(keySet) == 0 || keySet[upperKey] {
			out = append(out, upperKey+"="+val)
		} else {
			out = append(out, line)
		}
	}

	dst := opts.Dst
	if dst == "" {
		dst = opts.Src
	}

	f, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("writing %s: %w", dst, err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, l := range out {
		fmt.Fprintln(w, l)
	}
	return w.Flush()
}

func readUpperLines(path string) ([]string, error) {
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
