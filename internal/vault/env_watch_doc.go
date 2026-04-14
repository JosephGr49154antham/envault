/*
Package vault – Watch

# Overview

Watch polls a plain .env file at a configurable interval and fires a callback
whenever the file content changes (detected via SHA-256 digest comparison).

This is useful for development workflows where you want to automatically
re-push or re-encrypt a .env file after it is edited.

# Usage

	stop := make(chan struct{})
	err := vault.Watch(cfg, ".env", stop, vault.WatchOptions{
		Interval: 3 * time.Second,
		OnChange: func(path string) {
			fmt.Println("detected change in", path)
			// e.g. call vault.Push(...)
		},
	})

# Notes

  - The vault must be initialised before calling Watch.
  - The first read establishes the baseline hash; OnChange is NOT called on
    the initial read, only on subsequent changes.
  - Send to the stop channel (or close it) to terminate the watcher.
*/
package vault
