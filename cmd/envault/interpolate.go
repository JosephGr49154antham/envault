package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourorg/envault/internal/vault"
)

func runInterpolate(args []string) {
	fs := flag.NewFlagSet("interpolate", flag.ExitOnError)
	src := fs.String("src", "", "source .env file (default: vault plain file)")
	dst := fs.String("dst", "", "output file (default: <src>.interpolated<ext>)")
	overlay := fs.String("overlay", "", "overlay .env file whose values take precedence")
	strict := fs.Bool("strict", false, "fail if a referenced variable is undefined")
	_ = fs.Parse(args)

	cfg := vault.DefaultConfig()

	opts := vault.InterpolateOptions{
		Src:     *src,
		Dst:     *dst,
		Overlay: *overlay,
		Strict:  *strict,
	}

	if err := vault.Interpolate(cfg, opts); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	outPath := *dst
	if outPath == "" {
		outPath = "<src>.interpolated<ext>"
	}
	fmt.Printf("interpolated env written to %s\n", outPath)
}
