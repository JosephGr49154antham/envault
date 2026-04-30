package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourusername/envault/internal/vault"
)

func runSlice(args []string) {
	fs := flag.NewFlagSet("slice", flag.ExitOnError)
	src := fs.String("src", ".env", "source .env file")
	dst := fs.String("dst", "", "destination file (default: <src>.slice)")
	start := fs.Int("start", 1, "first key=value line to include (1-based)")
	end := fs.Int("end", 0, "last key=value line to include (0 = EOF)")
	force := fs.Bool("force", false, "overwrite destination if it exists")
	fs.Parse(args) //nolint:errcheck

	cfg := vault.DefaultConfig()

	err := vault.Slice(cfg, vault.SliceOptions{
		Src:   *src,
		Dst:   *dst,
		Start: *start,
		End:   *end,
		Force: *force,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	dstPath := *dst
	if dstPath == "" {
		dstPath = *src + ".slice"
	}
	fmt.Printf("sliced lines %d", *start)
	if *end > 0 {
		fmt.Printf("–%d", *end)
	}
	fmt.Printf(" → %s\n", dstPath)
}
