package vault

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// CastResult holds the result of casting a single env key's value.
type CastResult struct {
	Key      string
	Original string
	Cast     string
	Type     string
	OK       bool
}

// CastOptions controls Cast behaviour.
type CastOptions struct {
	Src  string
	Dst  string
	Keys []string // if empty, cast all keys
}

// Cast reads an env file and rewrites values into their inferred canonical
// types: booleans are normalised to true/false, integers are stripped of
// leading zeros, and floats are normalised. Strings are left untouched.
func Cast(cfg Config, opts CastOptions) ([]CastResult, error) {
	if !IsInitialised(cfg) {
		return nil, fmt.Errorf("vault not initialised")
	}

	src := opts.Src
	if src == "" {
		src = cfg.PlainFile
	}

	lines, err := readCastLines(src)
	if err != nil {
		return nil, err
	}

	keySet := make(map[string]bool, len(opts.Keys))
	for _, k := range opts.Keys {
		keySet[k] = true
	}

	var results []CastResult
	var out []string

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
		key := strings.TrimSpace(trimmed[:idx])
		val := strings.TrimSpace(trimmed[idx+1:])

		if len(keySet) > 0 && !keySet[key] {
			out = append(out, line)
			continue
		}

		casted, typ := castValue(val)
		results = append(results, CastResult{
			Key:      key,
			Original: val,
			Cast:     casted,
			Type:     typ,
			OK:       casted != val,
		})
		out = append(out, key+"="+casted)
	}

	dst := opts.Dst
	if dst == "" {
		dst = src
	}

	if err := os.WriteFile(dst, []byte(strings.Join(out, "\n")+"\n"), 0o600); err != nil {
		return nil, fmt.Errorf("cast: write %s: %w", dst, err)
	}
	return results, nil
}

func castValue(v string) (string, string) {
	unquoted := strings.Trim(v, `"`)
	lower := strings.ToLower(unquoted)
	switch lower {
	case "true", "yes", "1", "on":
		return "true", "bool"
	case "false", "no", "0", "off":
		return "false", "bool"
	}
	if i, err := strconv.ParseInt(unquoted, 10, 64); err == nil {
		return strconv.FormatInt(i, 10), "int"
	}
	if f, err := strconv.ParseFloat(unquoted, 64); err == nil {
		return strconv.FormatFloat(f, 'f', -1, 64), "float"
	}
	return v, "string"
}

func readCastLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cast: read %s: %w", path, err)
	}
	raw := strings.TrimRight(string(data), "\n")
	if raw == "" {
		return nil, nil
	}
	return strings.Split(raw, "\n"), nil
}
