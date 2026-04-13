package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// SchemaEntry describes a single expected environment variable.
type SchemaEntry struct {
	Key      string
	Required bool
	Comment  string
}

// SchemaResult holds the outcome of a schema validation.
type SchemaResult struct {
	Missing []string
	Extra   []string
}

// ValidateSchema checks the plain .env file against a .env.schema file.
// The schema file lists expected keys, one per line, with optional leading
// "#" comment lines. A key prefixed with "!" is required.
func ValidateSchema(cfg Config, schemaPath string) (*SchemaResult, error) {
	if !IsInitialised(cfg) {
		return nil, fmt.Errorf("vault not initialised")
	}

	schema, err := loadSchema(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("load schema: %w", err)
	}

	envKeys, err := loadEnvKeys(cfg.PlainFile)
	if err != nil {
		return nil, fmt.Errorf("load env file: %w", err)
	}

	envSet := make(map[string]struct{}, len(envKeys))
	for _, k := range envKeys {
		envSet[k] = struct{}{}
	}

	result := &SchemaResult{}
	schemaSet := make(map[string]struct{}, len(schema))
	for _, entry := range schema {
		schemaSet[entry.Key] = struct{}{}
		if entry.Required {
			if _, ok := envSet[entry.Key]; !ok {
				result.Missing = append(result.Missing, entry.Key)
			}
		}
	}

	for k := range envSet {
		if _, ok := schemaSet[k]; !ok {
			result.Extra = append(result.Extra, k)
		}
	}

	return result, nil
}

func loadSchema(path string) ([]SchemaEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []SchemaEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		required := false
		if strings.HasPrefix(line, "!") {
			required = true
			line = line[1:]
		}
		entries = append(entries, SchemaEntry{Key: strings.TrimSpace(line), Required: required})
	}
	return entries, scanner.Err()
}

func loadEnvKeys(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var keys []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) >= 1 {
			keys = append(keys, strings.TrimSpace(parts[0]))
		}
	}
	return keys, scanner.Err()
}
