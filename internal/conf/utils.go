package conf

import (
	"os"
	"os/user"
	"path/filepath"
)

// CurrentUsername returns the username of the current user.
func CurrentUsername() string {
	username := os.Getenv("USER")
	if len(username) > 0 {
		return username
	}

	username = os.Getenv("USERNAME")
	if len(username) > 0 {
		return username
	}

	if user, err := user.Current(); err == nil {
		username = user.Username
	}
	return username
}

// ensureAbs prepends the WorkDir to the given path if it is not an absolute path.
func ensureAbs(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(WorkDir(), path)
}

// CheckRunUser returns false if configured run user does not match actual user that
// runs the app. The first return value is the actual user name. This check is ignored
// under Windows since SSH remote login is not the main method to login on Windows.
func CheckRunUser(runUser string) (string, bool) {
	if IsWindowsRuntime() {
		return "", true
	}

	currentUser := CurrentUsername()
	return currentUser, runUser == currentUser
}
