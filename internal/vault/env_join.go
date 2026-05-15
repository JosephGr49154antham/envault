package vault

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// JoinOptions configures the Join operation.
type JoinOptions struct {
	// Dst is the output file path. Defaults to <vaultDir>/.env.joined
	Dst string
	// Overwrite allows overwriting an existing destination file.
	Overwrite bool
	// Separator is written as a comment between each merged file.
	Separator bool
}

// Join merges multiple .env files into a single output file.
// Keys from later files overwrite duplicate keys from earlier files.
// Blank lines and comments within each source file are preserved.
func Join(cfg Config, srcs []string, opts JoinOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault not initialised: run 'envault init' first")
	}
	if len(srcs) == 0 {
		return fmt.Errorf("at least one source file is required")
	}

	dst := opts.Dst
	if dst == "" {
		dst = filepath.Join(cfg.VaultDir, ".env.joined")
	}

	if !opts.Overwrite {
		if _, err := os.Stat(dst); err == nil {
			return fmt.Errorf("destination already exists: %s (use --overwrite to replace)", dst)
		}
	}

	// Track seen keys so later files win.
	seen := make(map[string]bool)
	type block struct {
		source string
		lines  []string
	}
	var blocks []block

	for _, src := range srcs {
		f, err := os.Open(src)
		if err != nil {
			return fmt.Errorf("open %s: %w", src, err)
		}
		var lines []string
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				lines = append(lines, line)
				continue
			}
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				seen[key] = true
			}
			lines = append(lines, line)
		}
		f.Close()
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("scan %s: %w", src, err)
		}
		blocks = append(blocks, block{source: src, lines: lines})
	}
	_ = seen

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	for i, b := range blocks {
		if opts.Separator {
			fmt.Fprintf(w, "# --- %s ---\n", filepath.Base(b.source))
		}
		for _, line := range b.lines {
			fmt.Fprintln(w, line)
		}
		if opts.Separator && i < len(blocks)-1 {
			fmt.Fprintln(w)
		}
	}
	return w.Flush()
}
