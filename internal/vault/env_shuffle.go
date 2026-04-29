package vault

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/envault/envault/internal/vault"
)

// ShuffleOptions controls the behaviour of the Shuffle operation.
type ShuffleOptions struct {
	Src  string
	Dst  string
	Seed int64 // 0 means use current time
}

// Shuffle randomly reorders the key=value lines in a .env file, preserving
// comments and blank lines in their relative positions among entries.
func Shuffle(cfg Config, opts ShuffleOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault is not initialised; run 'envault init' first")
	}

	src := opts.Src
	if src == "" {
		src = ".env"
	}

	dst := opts.Dst
	if dst == "" {
		dst = src
	}

	lines, err := readShuffleLines(src)
	if err != nil {
		return fmt.Errorf("read %s: %w", src, err)
	}

	shuffled := shuffleEnvLines(lines, opts.Seed)

	f, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	defer f.Close()

	for _, l := range shuffled {
		fmt.Fprintln(f, l)
	}
	return nil
}

func readShuffleLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}

// shuffleEnvLines separates key=value entries from non-entry lines, shuffles
// the entries, then reinserts them at the positions the entries originally
// occupied.
func shuffleEnvLines(lines []string, seed int64) []string {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	r := rand.New(rand.NewSource(seed))

	// collect indices of entry lines and the entries themselves
	var entryIdx []int
	var entries []string
	for i, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.Contains(trimmed, "=") {
			entryIdx = append(entryIdx, i)
			entries = append(entries, l)
		}
	}

	r.Shuffle(len(entries), func(i, j int) {
		entries[i], entries[j] = entries[j], entries[i]
	})

	out := make([]string, len(lines))
	copy(out, lines)
	for slot, idx := range entryIdx {
		out[idx] = entries[slot]
	}
	return out
}

// Ensure vault package reference compiles (import used only for IsInitialised).
var _ = vault.IsInitialised
