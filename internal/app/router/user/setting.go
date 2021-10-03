package user

import (
	// "bytes"
	// "encoding/base64"
	// "fmt"
	// "html/template"
	// "image/png"
	// "io/ioutil"
	// "strings"

	// "github.com/pquerna/otp"
	// "github.com/pquerna/otp/totp"

	"github.com/midoks/imail/internal/app/context"
	// "github.com/midoks/imail/internal/app/form"
	// "github.com/midoks/imail/internal/conf"
	// "github.com/midoks/imail/internal/db"
)

const (
	SETTINGS_PROFILE                   = "user/settings/profile"
	SETTINGS_AVATAR                    = "user/settings/avatar"
	SETTINGS_PASSWORD                  = "user/settings/password"
	SETTINGS_EMAILS                    = "user/settings/email"
	SETTINGS_SSH_KEYS                  = "user/settings/sshkeys"
	SETTINGS_SECURITY                  = "user/settings/security"
	SETTINGS_TWO_FACTOR_ENABLE         = "user/settings/two_factor_enable"
	SETTINGS_TWO_FACTOR_RECOVERY_CODES = "user/settings/two_factor_recovery_codes"
	SETTINGS_REPOSITORIES              = "user/settings/repositories"
	SETTINGS_ORGANIZATIONS             = "user/settings/organizations"
	SETTINGS_APPLICATIONS              = "user/settings/applications"
	SETTINGS_DELETE                    = "user/settings/delete"
	NOTIFICATION                       = "user/notification"
)

func Settings(c *context.Context) {
	c.Title("settings.profile")
	c.PageIs("SettingsProfile")
	// c.Data["origin_name"] = c.User.Name
	// c.Data["name"] = c.User.Name
	// c.Data["full_name"] = c.User.FullName
	// c.Data["email"] = c.User.Email
	// c.Data["website"] = c.User.Website
	// c.Data["location"] = c.User.Location
	c.Success(SETTINGS_PROFILE)
}
