package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/envault/internal/vault"
)

func runWatch(args []string) {
	fs := flag.NewFlagSet("watch", flag.ExitOnError)
	src := fs.String("src", ".env", "plain .env file to watch")
	interval := fs.Duration("interval", 2*time.Second, "poll interval (e.g. 1s, 500ms)")
	autoPush := fs.Bool("push", false, "automatically push (encrypt) on change")
	_ = fs.Parse(args)

	cfg := vault.DefaultConfig()

	if !vault.IsInitialised(cfg) {
		fmt.Fprintln(os.Stderr, "error: vault not initialised – run 'envault init' first")
		os.Exit(1)
	}

	fmt.Printf("watching %s (interval: %s)\n", *src, *interval)

	stop := make(chan struct{})
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		fmt.Println("\nstopping watcher")
		close(stop)
	}()

	opts := vault.WatchOptions{
		Interval: *interval,
		OnChange: func(path string) {
			fmt.Printf("change detected in %s\n", path)
			if *autoPush {
				if err := vault.Push(cfg, path); err != nil {
					fmt.Fprintf(os.Stderr, "auto-push failed: %v\n", err)
				} else {
					fmt.Println("auto-push complete")
				}
			}
		},
		OnError: func(path string, err error) {
			fmt.Fprintf(os.Stderr, "watch error (%s): %v\n", path, err)
		},
	}

	if err := vault.Watch(cfg, *src, stop, opts); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
