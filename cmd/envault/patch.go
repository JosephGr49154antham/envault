package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/yourorg/envault/internal/vault"
)

func runPatch(args []string) {
	fs := flag.NewFlagSet("patch", flag.ExitOnError)
	src := fs.String("src", "", "source .env file (default: vault plain file)")
	dst := fs.String("dst", "", "output file (default: same as src, in-place)")
	set := fs.String("set", "", "comma-separated KEY=VALUE pairs to upsert")
	del := fs.String("del", "", "comma-separated keys to delete")
	_ = fs.Parse(args)

	cfg := vault.DefaultConfig()

	opts := vault.PatchOptions{
		Src:     *src,
		Dst:     *dst,
		Upserts: make(map[string]string),
	}

	if *set != "" {
		for _, pair := range strings.Split(*set, ",") {
			pair = strings.TrimSpace(pair)
			if pair == "" {
				continue
			}
			idx := strings.IndexByte(pair, '=')
			if idx < 0 {
				fmt.Fprintf(os.Stderr, "patch: invalid pair %q (expected KEY=VALUE)\n", pair)
				os.Exit(1)
			}
			opts.Upserts[pair[:idx]] = pair[idx+1:]
		}
	}

	if *del != "" {
		for _, k := range strings.Split(*del, ",") {
			k = strings.TrimSpace(k)
			if k != "" {
				opts.Deletions = append(opts.Deletions, k)
			}
		}
	}

	if err := vault.Patch(cfg, opts); err != nil {
		fmt.Fprintf(os.Stderr, "patch: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("patch applied successfully")
}
