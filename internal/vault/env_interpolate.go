package vault

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// InterpolateOptions controls how Interpolate behaves.
type InterpolateOptions struct {
	Src     string // source .env file
	Dst     string // output file; defaults to <src>.interpolated
	Overlay string // optional second .env file whose values take precedence
	Strict  bool   // return an error if a referenced variable is undefined
}

var interpolateVarRe = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}|\$([A-Za-z_][A-Za-z0-9_]*)`)

// Interpolate expands ${VAR} / $VAR references inside .env values using the
// variables defined in the same file (and optionally an overlay file).
// The resolved file is written to Dst.
func Interpolate(cfg Config, opts InterpolateOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault not initialised — run `envault init` first")
	}

	src := opts.Src
	if src == "" {
		src = cfg.PlainFile
	}

	dst := opts.Dst
	if dst == "" {
		ext := filepath.Ext(src)
		dst = strings.TrimSuffix(src, ext) + ".interpolated" + ext
	}

	env, err := buildInterpolateMap(src, opts.Overlay)
	if err != nil {
		return err
	}

	lines, err := readInterpolateLines(src)
	if err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	for _, line := range lines {
		resolved, err := interpolateLine(line, env, opts.Strict)
		if err != nil {
			return err
		}
		fmt.Fprintln(w, resolved)
	}
	return w.Flush()
}

func buildInterpolateMap(src, overlay string) (map[string]string, error) {
	env := map[string]string{}
	for _, path := range []string{src, overlay} {
		if path == "" {
			continue
		}
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", path, err)
		}
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if k, v, ok := strings.Cut(line, "="); ok {
				env[strings.TrimSpace(k)] = strings.Trim(strings.TrimSpace(v), `"`)
			}
		}
		f.Close()
	}
	return env, nil
}

func readInterpolateLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}

func interpolateLine(line string, env map[string]string, strict bool) (string, error) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return line, nil
	}
	eq := strings.Index(line, "=")
	if eq < 0 {
		return line, nil
	}
	key := line[:eq]
	val := line[eq+1:]
	var expandErr error
	expanded := interpolateVarRe.ReplaceAllStringFunc(val, func(match string) string {
		name := interpolateVarRe.FindStringSubmatch(match)
		var varName string
		if name[1] != "" {
			varName = name[1]
		} else {
			varName = name[2]
		}
		if v, ok := env[varName]; ok {
			return v
		}
		if osVal := os.Getenv(varName); osVal != "" {
			return osVal
		}
		if strict {
			expandErr = fmt.Errorf("undefined variable: %s", varName)
		}
		return match
	})
	if expandErr != nil {
		return "", expandErr
	}
	return key + "=" + expanded, nil
}
