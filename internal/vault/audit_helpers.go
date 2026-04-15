package vault

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CurrentGitUser attempts to resolve the git user.email for the current
// repository as an audit actor identifier.
// Falls back to user.name, then to the OS username, then to an empty string
// if git is unavailable or not configured.
func CurrentGitUser() string {
	if email := gitConfig("user.email"); email != "" {
		return email
	}
	if name := gitConfig("user.name"); name != "" {
		return name
	}
	if osUser, err := osUsername(); err == nil {
		return osUser
	}
	return ""
}

// gitConfig reads a single git config key using the git CLI.
func gitConfig(key string) string {
	out, err := exec.Command("git", "config", "--get", key).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// osUsername returns the current OS user's login name via the USER or USERNAME
// environment variables, falling back to os.Hostname as a last resort.
func osUsername() (string, error) {
	if u := os.Getenv("USER"); u != "" {
		return u, nil
	}
	if u := os.Getenv("USERNAME"); u != "" {
		return u, nil
	}
	return "", fmt.Errorf("could not determine OS username")
}

// RecordEventWithGitUser is a convenience wrapper around RecordEvent that
// automatically resolves the current git user as the actor.
func RecordEventWithGitUser(cfg Config, op, details string) error {
	user := CurrentGitUser()
	if err := RecordEvent(cfg, op, user, details); err != nil {
		return fmt.Errorf("audit record (%s): %w", op, err)
	}
	return nil
}
