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
	"github.com/midoks/imail/internal/app/form"
	// "github.com/midoks/imail/internal/conf"
	// "github.com/midoks/imail/internal/db"
)

const (
	SETTINGS_PROFILE  = "user/settings/profile"
	SETTINGS_AVATAR   = "user/settings/avatar"
	SETTINGS_PASSWORD = "user/settings/password"
	SETTINGS_EMAILS   = "user/settings/email"
	SETTINGS_DELETE   = "user/settings/delete"
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

func SettingsPassword(c *context.Context) {
	c.Title("settings.password")
	c.PageIs("SettingsPassword")
	c.Success(SETTINGS_PASSWORD)
}

func SettingsPasswordPost(c *context.Context, f form.ChangePassword) {
	c.Title("settings.password")
	c.PageIs("SettingsPassword")

	if c.HasError() {
		c.Success(SETTINGS_PASSWORD)
		return
	}

	// if !c.User.ValidatePassword(f.OldPassword) {
	// 	c.Flash.Error(c.Tr("settings.password_incorrect"))
	// } else if f.Password != f.Retype {
	// 	c.Flash.Error(c.Tr("form.password_not_match"))
	// } else {
	// 	c.User.Passwd = f.Password
	// 	var err error
	// 	if c.User.Salt, err = db.GetUserSalt(); err != nil {
	// 		c.Errorf(err, "get user salt")
	// 		return
	// 	}
	// 	c.User.EncodePassword()
	// 	if err := db.UpdateUser(c.User); err != nil {
	// 		c.Errorf(err, "update user")
	// 		return
	// 	}
	// 	c.Flash.Success(c.Tr("settings.change_password_success"))
	// }

	c.RedirectSubpath("/user/settings/password")
}
