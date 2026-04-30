package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ReverseOptions controls the behaviour of the Reverse operation.
type ReverseOptions struct {
	Src string
	Dst string
	Overwrite bool
}

// Reverse reads a .env file and writes the entries in reversed order,
// preserving comments and blank lines in their relative positions.
func Reverse(cfg Config, opts ReverseOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault is not initialised: run 'envault init' first")
	}

	if opts.Dst == "" {
		opts.Dst = opts.Src + ".reversed"
	}

	if !opts.Overwrite {
		if _, err := os.Stat(opts.Dst); err == nil {
			return fmt.Errorf("destination already exists: %s (use --overwrite to replace)", opts.Dst)
		}
	}

	lines, err := readReverseLines(opts.Src)
	if err != nil {
		return fmt.Errorf("read %s: %w", opts.Src, err)
	}

	reversed := reverseEnvLines(lines)

	out, err := os.Create(opts.Dst)
	if err != nil {
		return fmt.Errorf("create %s: %w", opts.Dst, err)
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	for _, l := range reversed {
		fmt.Fprintln(w, l)
	}
	return w.Flush()
}

func readReverseLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// reverseEnvLines reverses only key=value entries; comments and blanks
// that appear between entries keep their relative order within the reversed
// entry list by being attached to the entry that follows them.
func reverseEnvLines(lines []string) []string {
	type block struct {
		header []string // comments/blanks before the entry
		entry  string
	}

	var blocks []block
	var pending []string

	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			pending = append(pending, l)
		} else {
			blocks = append(blocks, block{header: pending, entry: l})
			pending = nil
		}
	}

	// reverse blocks
	for i, j := 0, len(blocks)-1; i < j; i, j = i+1, j-1 {
		blocks[i], blocks[j] = blocks[j], blocks[i]
	}

	var result []string
	for _, b := range blocks {
		result = append(result, b.header...)
		result = append(result, b.entry)
	}
	result = append(result, pending...)
	return result
}
