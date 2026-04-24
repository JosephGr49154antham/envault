package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ResolveOptions controls the behaviour of Resolve.
type ResolveOptions struct {
	Src    string // source .env file (default: vault plain file)
	Dst    string // output file (default: overwrite src)
	Strict bool   // return an error if any ${VAR} reference is unresolvable
}

// Resolve expands variable references (${VAR} or $VAR) inside a .env file
// using values defined earlier in the same file and from the current process
// environment. References that cannot be resolved are left intact unless
// Strict mode is enabled.
func Resolve(cfg Config, opts ResolveOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault not initialised: run 'envault init' first")
	}

	src := opts.Src
	if src == "" {
		src = cfg.PlainFile
	}

	dst := opts.Dst
	if dst == "" {
		dst = src
	}

	lines, err := readResolveLines(src)
	if err != nil {
		return fmt.Errorf("resolve: read %s: %w", src, err)
	}

	resolved, err := resolveLines(lines, opts.Strict)
	if err != nil {
		return err
	}

	if err := writeResolveLines(dst, resolved); err != nil {
		return fmt.Errorf("resolve: write %s: %w", dst, err)
	}
	return nil
}

func readResolveLines(path string) ([]string, error) {
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

func resolveLines(lines []string, strict bool) ([]string, error) {
	// Build a local lookup from already-seen assignments.
	local := make(map[string]string)

	out := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			out = append(out, line)
			continue
		}

		eqIdx := strings.IndexByte(trimmed, '=')
		if eqIdx < 0 {
			out = append(out, line)
			continue
		}

		key := strings.TrimSpace(trimmed[:eqIdx])
		rawVal := trimmed[eqIdx+1:]

		expanded, err := expandValue(rawVal, local, strict)
		if err != nil {
			return nil, fmt.Errorf("resolve: key %q: %w", key, err)
		}

		local[key] = expanded
		out = append(out, key+"="+expanded)
	}
	return out, nil
}

func expandValue(val string, local map[string]string, strict bool) (string, error) {
	var sb strings.Builder
	i := 0
	for i < len(val) {
		if val[i] != '$' {
			sb.WriteByte(val[i])
			i++
			continue
		}
		// Found '$'
		i++
		if i >= len(val) {
			sb.WriteByte('$')
			break
		}

		var varName string
		if val[i] == '{' {
			// ${VAR}
			end := strings.IndexByte(val[i:], '}')
			if end < 0 {
				sb.WriteByte('$')
				continue
			}
			varName = val[i+1 : i+end]
			i += end + 1
		} else {
			// $VAR
			j := i
			for j < len(val) && isVarChar(val[j]) {
				j++
			}
			varName = val[i:j]
			i = j
		}

		if varName == "" {
			sb.WriteByte('$')
			continue
		}

		if v, ok := local[varName]; ok {
			sb.WriteString(v)
		} else if v := os.Getenv(varName); v != "" {
			sb.WriteString(v)
		} else if strict {
			return "", fmt.Errorf("unresolvable reference: $%s", varName)
		} else {
			// Leave the reference intact.
			sb.WriteString("${" + varName + "}")
		}
	}
	return sb.String(), nil
}

func isVarChar(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
		(c >= '0' && c <= '9') || c == '_'
}

func writeResolveLines(path string, lines []string) error {
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
