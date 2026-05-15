package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// IntersectOptions controls the behaviour of the Intersect operation.
type IntersectOptions struct {
	// Src is the primary .env file.
	Src string
	// With is the second .env file to intersect against.
	With string
	// Dst is the output file. Defaults to <Src>.intersect.env.
	Dst string
	// Overwrite allows overwriting an existing destination file.
	Overwrite bool
	// ValuesFrom selects which file's values are written to the output.
	// "src" (default) keeps values from Src; "with" keeps values from With.
	ValuesFrom string
}

// Intersect writes only the keys that appear in both Src and With to Dst,
// preserving the value from the file specified by ValuesFrom (default: "src").
// Comments and blank lines from Src are preserved when their surrounding keys
// survive the intersection.
func Intersect(cfg Config, opts IntersectOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault is not initialised; run 'envault init' first")
	}

	if opts.Src == "" {
		opts.Src = ".env"
	}
	if opts.With == "" {
		return fmt.Errorf("--with file must be specified")
	}
	if opts.Dst == "" {
		opts.Dst = intersectDefaultDst(opts.Src)
	}
	if opts.ValuesFrom == "" {
		opts.ValuesFrom = "src"
	}
	if opts.ValuesFrom != "src" && opts.ValuesFrom != "with" {
		return fmt.Errorf("--values-from must be \"src\" or \"with\"")
	}

	// Build the key set from the "with" file.
	withKeys, withValues, err := readIntersectMap(opts.With)
	if err != nil {
		return fmt.Errorf("reading --with file: %w", err)
	}

	if !opts.Overwrite {
		if _, err := os.Stat(opts.Dst); err == nil {
			return fmt.Errorf("destination %q already exists; use --overwrite to replace it", opts.Dst)
		}
	}

	srcFile, err := os.Open(opts.Src)
	if err != nil {
		return fmt.Errorf("opening source file: %w", err)
	}
	defer srcFile.Close()

	var out []string
	scanner := bufio.NewScanner(srcFile)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Always keep comments and blank lines.
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			out = append(out, line)
			continue
		}

		eqIdx := strings.Index(trimmed, "=")
		if eqIdx < 0 {
			// Not a valid key=value line; skip.
			continue
		}

		key := strings.TrimSpace(trimmed[:eqIdx])
		if _, ok := withKeys[key]; !ok {
			// Key not present in "with" file — exclude.
			continue
		}

		if opts.ValuesFrom == "with" {
			out = append(out, key+"="+withValues[key])
		} else {
			out = append(out, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading source file: %w", err)
	}

	result := strings.Join(out, "\n")
	if len(out) > 0 {
		result += "\n"
	}

	if err := os.WriteFile(opts.Dst, []byte(result), 0o644); err != nil {
		return fmt.Errorf("writing destination file: %w", err)
	}
	return nil
}

// readIntersectMap returns a set of keys and a key→value map for the given file.
func readIntersectMap(path string) (map[string]struct{}, map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	keys := make(map[string]struct{})
	values := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		eqIdx := strings.Index(line, "=")
		if eqIdx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:eqIdx])
		val := line[eqIdx+1:]
		keys[key] = struct{}{}
		values[key] = val
	}
	return keys, values, scanner.Err()
}

// intersectDefaultDst derives a default output path from the source file name.
func intersectDefaultDst(src string) string {
	if src == ".env" {
		return ".env.intersect"
	}
	return src + ".intersect"
}
