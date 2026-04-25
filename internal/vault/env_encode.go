package vault

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EncodeOptions controls how Encode behaves.
type EncodeOptions struct {
	// Src is the source .env file to encode.
	Src string
	// Dst is the destination file. If empty, defaults to <src>.encoded.
	Dst string
	// Encoding selects the encoding scheme: "base64" (default) or "hex".
	Encoding string
	// Overwrite allows an existing destination file to be replaced.
	Overwrite bool
	// KeysOnly encodes only the values, leaving keys in plain text.
	KeysOnly bool
}

// DecodeOptions controls how Decode behaves.
type DecodeOptions struct {
	// Src is the source encoded .env file.
	Src string
	// Dst is the destination file. If empty, defaults to <src>.decoded.
	Dst string
	// Encoding selects the encoding scheme: "base64" (default) or "hex".
	Encoding string
	// Overwrite allows an existing destination file to be replaced.
	Overwrite bool
}

// Encode reads a .env file and writes a copy where every value is encoded
// using the selected scheme. Comments and blank lines are preserved as-is.
func Encode(cfg Config, opts EncodeOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault not initialised; run 'envault init' first")
	}

	if opts.Encoding == "" {
		opts.Encoding = "base64"
	}

	if opts.Dst == "" {
		opts.Dst = opts.Src + ".encoded"
	}

	if !opts.Overwrite {
		if _, err := os.Stat(opts.Dst); err == nil {
			return fmt.Errorf("destination file already exists: %s (use --overwrite to replace)", opts.Dst)
		}
	}

	lines, err := readEncodeLines(opts.Src)
	if err != nil {
		return fmt.Errorf("reading source file: %w", err)
	}

	encoded := make([]string, 0, len(lines))
	for _, line := range lines {
		encoded = append(encoded, encodeEnvLine(line, opts.Encoding, opts.KeysOnly))
	}

	if err := os.MkdirAll(filepath.Dir(opts.Dst), 0o700); err != nil {
		return fmt.Errorf("creating destination directory: %w", err)
	}

	return os.WriteFile(opts.Dst, []byte(strings.Join(encoded, "\n")+"\n"), 0o600)
}

// Decode reads an encoded .env file produced by Encode and writes the
// decoded values back to a plain .env file.
func Decode(cfg Config, opts DecodeOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault not initialised; run 'envault init' first")
	}

	if opts.Encoding == "" {
		opts.Encoding = "base64"
	}

	if opts.Dst == "" {
		// Strip a trailing .encoded suffix if present, otherwise append .decoded.
		base := strings.TrimSuffix(opts.Src, ".encoded")
		if base == opts.Src {
			base = opts.Src + ".decoded"
		}
		opts.Dst = base
	}

	if !opts.Overwrite {
		if _, err := os.Stat(opts.Dst); err == nil {
			return fmt.Errorf("destination file already exists: %s (use --overwrite to replace)", opts.Dst)
		}
	}

	lines, err := readEncodeLines(opts.Src)
	if err != nil {
		return fmt.Errorf("reading source file: %w", err)
	}

	decoded := make([]string, 0, len(lines))
	for _, line := range lines {
		dl, err := decodeEnvLine(line, opts.Encoding)
		if err != nil {
			return fmt.Errorf("decoding line %q: %w", line, err)
		}
		decoded = append(decoded, dl)
	}

	if err := os.MkdirAll(filepath.Dir(opts.Dst), 0o700); err != nil {
		return fmt.Errorf("creating destination directory: %w", err)
	}

	return os.WriteFile(opts.Dst, []byte(strings.Join(decoded, "\n")+"\n"), 0o600)
}

// readEncodeLines reads all lines from path, stripping the trailing newline
// from each so callers work with clean strings.
func readEncodeLines(path string) ([]string, error) {
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

// encodeEnvLine encodes the value portion of a KEY=VALUE line. Comments and
// blank lines are returned unchanged.
func encodeEnvLine(line, encoding string, keysOnly bool) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return line
	}

	eq := strings.IndexByte(line, '=')
	if eq < 0 {
		return line
	}

	key := line[:eq]
	value := line[eq+1:]

	encodedValue := encodeString(value, encoding)

	if keysOnly {
		return key + "=" + encodedValue
	}

	encodedKey := encodeString(key, encoding)
	return encodedKey + "=" + encodedValue
}

// decodeEnvLine decodes the value (and key when present) of a KEY=VALUE line.
func decodeEnvLine(line, encoding string) (string, error) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return line, nil
	}

	eq := strings.IndexByte(line, '=')
	if eq < 0 {
		return line, nil
	}

	encodedKey := line[:eq]
	encodedValue := line[eq+1:]

	decodedKey, err := decodeString(encodedKey, encoding)
	if err != nil {
		// Key may not have been encoded (KeysOnly mode); fall back to raw.
		decodedKey = encodedKey
	}

	decodedValue, err := decodeString(encodedValue, encoding)
	if err != nil {
		return "", fmt.Errorf("decoding value for key %q: %w", decodedKey, err)
	}

	return decodedKey + "=" + decodedValue, nil
}

func encodeString(s, encoding string) string {
	switch encoding {
	case "hex":
		return fmt.Sprintf("%x", s)
	default: // base64
		return base64.StdEncoding.EncodeToString([]byte(s))
	}
}

func decodeString(s, encoding string) (string, error) {
	switch encoding {
	case "hex":
		var buf []byte
		_, err := fmt.Sscanf(s, "%x", &buf)
		if err != nil {
			return "", err
		}
		return string(buf), nil
	default: // base64
		b, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
}
