package conf

import (
	"github.com/midoks/imail/internal/tools"
	"path/filepath"
)

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

	currentUser := tools.CurrentUsername()
	return currentUser, runUser == currentUser
}
