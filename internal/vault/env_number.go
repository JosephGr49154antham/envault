package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// NumberOptions controls how line numbers are rendered.
type NumberOptions struct {
	Src    string
	Dst    string
	OnlyKV bool // number only key=value lines, skip comments and blanks
}

// Number writes a copy of the .env file with line numbers prepended to each
// line. If Dst is empty the output is written to <src>.numbered.
func Number(cfg Config, opts NumberOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault not initialised: run 'envault init' first")
	}

	f, err := os.Open(opts.Src)
	if err != nil {
		return fmt.Errorf("open %s: %w", opts.Src, err)
	}
	defer f.Close()

	dst := opts.Dst
	if dst == "" {
		dst = opts.Src + ".numbered"
	}

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	defer out.Close()

	scanner := bufio.NewScanner(f)
	lineNum := 1
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		isKV := trimmed != "" && !strings.HasPrefix(trimmed, "#")

		if opts.OnlyKV && !isKV {
			fmt.Fprintln(out, line)
		} else {
			fmt.Fprintf(out, "%4d  %s\n", lineNum, line)
			lineNum++
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read %s: %w", opts.Src, err)
	}
	return nil
}
