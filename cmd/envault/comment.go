package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/nicholasgasior/envault/internal/vault"
)

func runComment(args []string) {
	fs := flag.NewFlagSet("comment", flag.ExitOnError)
	src := fs.String("src", ".env", "source .env file")
	dst := fs.String("dst", "", "destination file (default: <src>.commented.env)")
	keys := fs.String("keys", "", "comma-separated list of keys to target (empty = all)")
	uncomment := fs.Bool("uncomment", false, "remove leading '#' instead of adding it")
	overwrite := fs.Bool("overwrite", false, "allow overwriting existing destination file")
	_ = fs.Parse(args)

	var keyList []string
	if *keys != "" {
		for _, k := range strings.Split(*keys, ",") {
			if t := strings.TrimSpace(k); t != "" {
				keyList = append(keyList, t)
			}
		}
	}

	cfg := vault.DefaultConfig()
	err := vault.Comment(cfg, vault.CommentOptions{
		Src:       *src,
		Dst:       *dst,
		Keys:      keyList,
		Uncomment: *uncomment,
		Overwrite: *overwrite,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	action := "commented"
	if *uncomment {
		action = "uncommented"
	}
	target := *dst
	if target == "" {
		target = strings.TrimSuffix(*src, ".env") + ".commented.env"
		if *uncomment {
			target = *src
		}
	}
	fmt.Printf("%s → %s\n", action, target)
}
