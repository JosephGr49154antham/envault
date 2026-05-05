package vault

// Comment adds or removes inline comments on key-value lines in a .env file.
//
// Usage:
//
//	Comment(cfg, CommentOptions{
//		Src:       ".env",
//		Dst:       ".env.out",       // optional; defaults to <src>.commented.env
//		Keys:      []string{"SECRET"}, // empty = all KV lines
//		Uncomment: false,             // true to strip leading '#'
//		Overwrite: false,
//	})
//
// Behaviour:
//   - When Uncomment is false (default), every matching key=value line is
//     prefixed with "# ", effectively disabling it.
//   - When Uncomment is true, lines that begin with "# KEY=..." are restored
//     to active key=value lines.
//   - Lines that are already comments, blank lines, and non-matching keys are
//     passed through unchanged.
//   - If Keys is empty, every key=value line is affected.
//   - The vault must be initialised before calling Comment.
