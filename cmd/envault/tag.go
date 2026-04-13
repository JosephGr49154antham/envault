package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/yourusername/envault/internal/vault"
)

func runTag(args []string) {
	fs := flag.NewFlagSet("tag", flag.ExitOnError)
	delFlag := fs.Bool("delete", false, "delete an existing tag")
	msgFlag := fs.String("m", "", "optional message for the tag")
	_ = fs.Parse(args)

	cfg := vault.DefaultConfig()

	// List tags when no positional args are given.
	if fs.NArg() == 0 {
		tags, err := vault.ListTags(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if len(tags) == 0 {
			fmt.Println("no tags found")
			return
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tSNAPSHOT\tCREATED\tMESSAGE")
		for _, t := range tags {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				t.Name,
				t.SnapshotID,
				t.CreatedAt.Format("2006-01-02 15:04"),
				t.Message,
			)
		}
		w.Flush()
		return
	}

	name := fs.Arg(0)

	if *delFlag {
		if err := vault.DeleteTag(cfg, name); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("tag %q deleted\n", name)
		return
	}

	if fs.NArg() < 2 {
		fmt.Fprintln(os.Stderr, "usage: envault tag <name> <snapshot-id> [-m message]")
		os.Exit(1)
	}
	snapshotID := fs.Arg(1)

	if err := vault.CreateTag(cfg, name, snapshotID, *msgFlag); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("tag %q -> %s\n", name, snapshotID)
}
