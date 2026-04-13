/*
Package vault — Schema Validation

# Overview

The schema validation feature allows teams to declare which environment
variables are expected in a .env file, and whether each one is required.

# Schema File Format

The schema file (typically .env.schema) contains one key per line:

  - Lines starting with '#' are treated as comments and ignored.
  - Blank lines are ignored.
  - A key prefixed with '!' is marked as required — its absence is reported
    as a missing key.
  - A key without '!' is optional — it is recognised but not enforced.

Example .env.schema:

	# Database
	!DB_HOST
	!DB_PORT
	DB_NAME

	# Application
	!APP_KEY
	APP_DEBUG

# Usage

	res, err := vault.ValidateSchema(cfg, ".env.schema")
	if err != nil {
	    log.Fatal(err)
	}
	for _, k := range res.Missing {
	    fmt.Println("missing required key:", k)
	}
	for _, k := range res.Extra {
	    fmt.Println("undeclared key:", k)
	}
*/
package vault
