package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// EchoOptions controls the output format of Echo.
type EchoOptions struct {
	Src      string
	Keys     []string
	Export   bool
	Quote    bool
	NullSep  bool
}

// Echo prints the resolved values of the given keys (or all keys) from src
// to stdout. Useful for quick inspection or shell eval pipelines.
func Echo(cfg Config, opts EchoOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault not initialised: run 'envault init' first")
	}

	src := opts.Src
	if src == "" {
		src = cfg.PlainFile
	}

	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer f.Close()

	wantAll := len(opts.Keys) == 0
	wantSet := make(map[string]struct{}, len(opts.Keys))
	for _, k := range opts.Keys {
		wantSet[k] = struct{}{}
	}

	sep := "\n"
	if opts.NullSep {
		sep = "\x00"
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		idx := strings.IndexByte(trimmed, '=')
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(trimmed[:idx])
		val := strings.TrimSpace(trimmed[idx+1:])
		val = strings.Trim(val, `"`)

		if !wantAll {
			if _, ok := wantSet[key]; !ok {
				continue
			}
		}

		if opts.Quote {
			val = `"` + val + `"`
		}

		if opts.Export {
			fmt.Printf("export %s=%s%s", key, val, sep)
		} else {
			fmt.Printf("%s=%s%s", key, val, sep)
		}
	}
	return scanner.Err()
}
