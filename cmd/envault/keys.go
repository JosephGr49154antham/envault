package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/example/envault/internal/vault"
)

func runKeys(args []string) {
	var (
		src        = ".env"
		sorted     bool
		valuesOnly bool
	)

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--src":
			i++
			if i < len(args) {
				src = args[i]
			}
		case "--sorted", "-s":
			sorted = true
		case "--values", "-v":
			valuesOnly = true
		}
	}

	if !filepath.IsAbs(src) {
		cwd, _ := os.Getwd()
		src = filepath.Join(cwd, src)
	}

	cfg := vault.DefaultConfig()
	results, err := vault.Keys(cfg, src, vault.KeysOptions{
		Sorted:     sorted,
		ValuesOnly: valuesOnly,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	for _, item := range results {
		fmt.Println(item)
	}
}
