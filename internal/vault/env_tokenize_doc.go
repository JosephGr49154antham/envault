/*
Package vault — Tokenize

# Overview

Tokenize parses a .env file into a structured list of Tokens.
Each token carries its line number, kind, key, value, and the
original raw text so callers can reconstruct or analyse the file
without losing formatting context.

# Token kinds

  - "key"     — a valid KEY=VALUE assignment
  - "comment" — a line starting with '#'
  - "blank"   — an empty or whitespace-only line
  - "invalid" — a line that could not be parsed

# Usage

	result, err := vault.Tokenize(cfg, ".env")
	if err != nil {
		log.Fatal(err)
	}
	for _, tok := range result.Tokens {
		fmt.Printf("[%s] line %d: %s\n", tok.Kind, tok.Line, tok.Raw)
	}
	if len(result.Errors) > 0 {
		fmt.Println("parse errors:", result.Errors)
	}
*/
package vault
