package router

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"

	"github.com/midoks/imail/internal/app/context"
	"github.com/midoks/imail/internal/app/form"
	"github.com/midoks/imail/internal/conf"
	"github.com/midoks/imail/internal/log"
	"github.com/midoks/imail/internal/tools"
)

const (
	INSTALL = "install"
)

func GlobalInit(customConf string) error {
	err := conf.Init(customConf)
	fmt.Println(err)
	return nil
}

func InstallInit(c *context.Context) {
	if conf.Security.InstallLock {
		c.NotFound()
		return
	}

	c.Title("install.install")
	c.PageIs("Install")

	c.Data["DbOptions"] = []string{"MySQL", "PostgreSQL", "SQLite3"}
}

func Install(c *context.Context) {
	f := form.Install{}

	// Database settings
	f.DbHost = conf.Database.Host
	f.DbUser = conf.Database.User
	f.DbName = conf.Database.Name
	f.DbPath = conf.Database.Path

	c.Data["CurDbOption"] = "PostgreSQL"
	switch conf.Database.Type {
	case "mysql":
		c.Data["CurDbOption"] = "MySQL"
	case "sqlite3":
		c.Data["CurDbOption"] = "SQLite3"
	}

	// Application general settings
	f.AppName = conf.App.BrandName

	// Note(unknwon): it's hard for Windows users change a running user,
	// 	so just use current one if config says default.
	if conf.IsWindowsRuntime() && conf.App.RunUser == "git" {
		f.RunUser = tools.CurrentUsername()
	} else {
		f.RunUser = conf.App.RunUser
	}

	f.Domain = conf.Web.Domain
	f.HttpPort = conf.Web.HttpPort
	f.LogRootPath = fmt.Sprintf("%s/%s", conf.WorkDir(), conf.Log.RootPath)

	// Server and other services settings
	form.Assign(f, c.Data)
	c.Success(INSTALL)
}

func InstallPost(c *context.Context, f form.Install) {
	c.Data["CurDbOption"] = f.DbType

	if c.HasError() {
		if c.HasValue("Err_AdminName") ||
			c.HasValue("Err_AdminPasswd") ||
			c.HasValue("Err_AdminEmail") {
			c.FormErr("Admin")
		}

		c.Success(INSTALL)
		return
	}

	if _, err := exec.LookPath("git"); err != nil {
		c.RenderWithErr(c.Tr("install.test_git_failed", err), INSTALL, &f)
		return
	}

	// Pass basic check, now test configuration.
	// Test database setting.
	dbTypes := map[string]string{
		"PostgreSQL": "postgres",
		"MySQL":      "mysql",
		"SQLite3":    "sqlite3",
	}
	conf.Database.Type = dbTypes[f.DbType]
	conf.Database.Host = f.DbHost
	conf.Database.User = f.DbUser
	conf.Database.Password = f.DbPasswd
	conf.Database.Name = f.DbName
	conf.Database.SSLMode = f.SSLMode
	conf.Database.Path = f.DbPath

	if conf.Database.Type == "sqlite3" && len(conf.Database.Path) == 0 {
		c.FormErr("DbPath")
		c.RenderWithErr(c.Tr("install.err_empty_db_path"), INSTALL, &f)
		return
	}

	// Set test engine.
	// if err := db.NewTestEngine(); err != nil {
	// 	if strings.Contains(err.Error(), `Unknown database type: sqlite3`) {
	// 		c.FormErr("DbType")
	// 		c.RenderWithErr(c.Tr("install.sqlite3_not_available", "https://gogs.io/docs/installation/install_from_binary.html"), INSTALL, &f)
	// 	} else {
	// 		c.FormErr("DbSetting")
	// 		c.RenderWithErr(c.Tr("install.invalid_db_setting", err), INSTALL, &f)
	// 	}
	// 	return
	// }

	// Test log root path.
	f.LogRootPath = strings.Replace(f.LogRootPath, "\\", "/", -1)
	if err := os.MkdirAll(f.LogRootPath, os.ModePerm); err != nil {
		c.FormErr("LogRootPath")
		c.RenderWithErr(c.Tr("install.invalid_log_root_path", err), INSTALL, &f)
		return
	}

	currentUser, match := conf.CheckRunUser(f.RunUser)
	if !match {
		c.FormErr("RunUser")
		c.RenderWithErr(c.Tr("install.run_user_not_match", f.RunUser, currentUser), INSTALL, &f)
		return
	}

	// Check logic loophole between disable self-registration and no admin account.
	if f.DisableRegistration && len(f.AdminName) == 0 {
		c.FormErr("Services", "Admin")
		c.RenderWithErr(c.Tr("install.no_admin_and_disable_registration"), INSTALL, f)
		return
	}

	// Check admin password.
	if len(f.AdminName) > 0 && len(f.AdminPasswd) == 0 {
		c.FormErr("Admin", "AdminPasswd")
		c.RenderWithErr(c.Tr("install.err_empty_admin_password"), INSTALL, f)
		return
	}
	if f.AdminPasswd != f.AdminConfirmPasswd {
		c.FormErr("Admin", "AdminPasswd")
		c.RenderWithErr(c.Tr("form.password_not_match"), INSTALL, f)
		return
	}

	// Save settings.
	cfg := ini.Empty()
	if tools.IsFile(conf.CustomConf) {
		// Keeps custom settings if there is already something.
		if err := cfg.Append(conf.CustomConf); err != nil {
			log.Error("Failed to load custom conf %q: %v", conf.CustomConf, err)
		}
	}

	cfg.Section("").Key("brand_name").SetValue(f.AppName)
	cfg.Section("").Key("run_user").SetValue(f.RunUser)
	cfg.Section("").Key("run_mode").SetValue("prod")

	cfg.Section("database").Key("type").SetValue(conf.Database.Type)
	cfg.Section("database").Key("host").SetValue(conf.Database.Host)
	cfg.Section("database").Key("name").SetValue(conf.Database.Name)
	cfg.Section("database").Key("user").SetValue(conf.Database.User)
	cfg.Section("database").Key("password").SetValue(conf.Database.Password)
	cfg.Section("database").Key("ssl_mode").SetValue(conf.Database.SSLMode)
	cfg.Section("database").Key("path").SetValue(conf.Database.Path)

	cfg.Section("web").Key("domain").SetValue(f.Domain)
	cfg.Section("web").Key("http_port").SetValue(f.HttpPort)

	cfg.Section("session").Key("provider").SetValue("file")

	mode := "file"
	if f.EnableConsoleMode {
		mode = "console, file"
	}
	cfg.Section("log").Key("mode").SetValue(mode)
	cfg.Section("log").Key("level").SetValue("Info")
	cfg.Section("log").Key("root_path").SetValue(f.LogRootPath)

	cfg.Section("security").Key("install_lock").SetValue("true")
	secretKey := tools.RandString(15)
	cfg.Section("security").Key("secret_key").SetValue(secretKey)

	_ = os.MkdirAll(filepath.Dir(conf.CustomConf), os.ModePerm)
	if err := cfg.SaveTo(conf.CustomConf); err != nil {
		c.RenderWithErr(c.Tr("install.save_config_failed", err), INSTALL, &f)
		return
	}

	// NOTE: We reuse the current value because this handler does not have access to CLI flags.
	err := GlobalInit(conf.CustomConf)
	if err != nil {
		c.RenderWithErr(c.Tr("install.init_failed", err), INSTALL, &f)
		return
	}

	// // Create admin account
	// if len(f.AdminName) > 0 {
	// 	u := &db.User{
	// 		Name:     f.AdminName,
	// 		Email:    f.AdminEmail,
	// 		Passwd:   f.AdminPasswd,
	// 		IsAdmin:  true,
	// 		IsActive: true,
	// 	}
	// 	if err := db.CreateUser(u); err != nil {
	// 		if !db.IsErrUserAlreadyExist(err) {
	// 			conf.Security.InstallLock = false
	// 			c.FormErr("AdminName", "AdminEmail")
	// 			c.RenderWithErr(c.Tr("install.invalid_admin_setting", err), INSTALL, &f)
	// 			return
	// 		}
	// 		log.Info("Admin account already exist")
	// 		u, _ = db.GetUserByName(u.Name)
	// 	}

	// 	// Auto-login for admin
	// 	_ = c.Session.Set("uid", u.ID)
	// 	_ = c.Session.Set("uname", u.Name)
	// }

	log.Info("first-time run install finished!")
	c.Flash.Success(c.Tr("install.install_success"))
	c.Redirect("/user/login")
}
