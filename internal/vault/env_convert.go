package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ConvertFormat describes the output format for Convert.
type ConvertFormat string

const (
	FormatDotenv ConvertFormat = "dotenv"
	FormatExport ConvertFormat = "export"
	FormatJSON   ConvertFormat = "json"
	FormatYAML   ConvertFormat = "yaml"
)

// ConvertOptions controls the behaviour of Convert.
type ConvertOptions struct {
	Src    string
	Dst    string
	Format ConvertFormat
}

// Convert reads a .env file and writes it in the requested format.
// If Dst is empty a default path is derived from Src with an appropriate extension.
func Convert(cfg Config, opts ConvertOptions) (string, error) {
	if !IsInitialised(cfg) {
		return "", fmt.Errorf("vault not initialised: run 'envault init' first")
	}

	if opts.Dst == "" {
		ext := formatExtension(opts.Format)
		base := strings.TrimSuffix(filepath.Base(opts.Src), filepath.Ext(opts.Src))
		opts.Dst = filepath.Join(filepath.Dir(opts.Src), base+ext)
	}

	if _, err := os.Stat(opts.Dst); err == nil {
		return "", fmt.Errorf("destination already exists: %s", opts.Dst)
	}

	pairs, err := readEnvFileToMap(opts.Src)
	if err != nil {
		return "", fmt.Errorf("reading source: %w", err)
	}

	var out string
	switch opts.Format {
	case FormatExport:
		out = renderExport(pairs)
	case FormatJSON:
		out = renderJSON(pairs)
	case FormatYAML:
		out = renderYAML(pairs)
	default:
		out = renderDotenv(pairs)
	}

	if err := os.WriteFile(opts.Dst, []byte(out), 0o600); err != nil {
		return "", fmt.Errorf("writing output: %w", err)
	}
	return opts.Dst, nil
}

func formatExtension(f ConvertFormat) string {
	switch f {
	case FormatJSON:
		return ".json"
	case FormatYAML:
		return ".yaml"
	default:
		return ".env"
	}
}

func renderDotenv(pairs map[string]string) string {
	var sb strings.Builder
	for k, v := range pairs {
		fmt.Fprintf(&sb, "%s=%s\n", k, v)
	}
	return sb.String()
}

func renderExport(pairs map[string]string) string {
	var sb strings.Builder
	for k, v := range pairs {
		fmt.Fprintf(&sb, "export %s=%q\n", k, v)
	}
	return sb.String()
}

func renderJSON(pairs map[string]string) string {
	var sb strings.Builder
	sb.WriteString("{\n")
	i := 0
	for k, v := range pairs {
		comma := ","
		if i == len(pairs)-1 {
			comma = ""
		}
		fmt.Fprintf(&sb, "  %q: %q%s\n", k, v, comma)
		i++
	}
	sb.WriteString("}\n")
	return sb.String()
}

func renderYAML(pairs map[string]string) string {
	var sb strings.Builder
	for k, v := range pairs {
		fmt.Fprintf(&sb, "%s: %q\n", k, v)
	}
	return sb.String()
}
