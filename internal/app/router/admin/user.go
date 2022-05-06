package admin

import (
	// "strings"

	"github.com/midoks/imail/internal/app/context"
	"github.com/midoks/imail/internal/app/form"
	"github.com/midoks/imail/internal/conf"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/log"
	"github.com/midoks/imail/internal/tools/paginater"
)

const (
	USERS     = "admin/user/list"
	USER_NEW  = "admin/user/new"
	USER_EDIT = "admin/user/edit"
)

func RenderUserSearch(c *context.Context, opts *db.UserSearchOptions) {
	page := c.QueryInt("page")
	if page <= 1 {
		page = 1
	}

	var (
		users []db.User
		count int64
		err   error
	)

	keyword := c.Query("q")
	if len(keyword) == 0 {
		users, err = db.UsersList(page, opts.PageSize)
		count = db.UsersCount()
	} else {
		users, count, err = db.UserSearchByName(&db.UserSearchOptions{
			Keyword:  keyword,
			OrderBy:  opts.OrderBy,
			Page:     page,
			PageSize: opts.PageSize,
		})
		if err != nil {
			c.Error(err, "search user by name")
			return
		}
	}
	c.Data["Keyword"] = keyword
	c.Data["Total"] = count
	c.Data["Page"] = paginater.New(int(count), opts.PageSize, page, 5)
	c.Data["Users"] = users

	c.Success(opts.TplName)
}

func Users(c *context.Context) {
	c.Data["Title"] = c.Tr("admin.users")
	c.Data["PageIsAdmin"] = true
	c.Data["PageIsAdminUsers"] = true

	RenderUserSearch(c, &db.UserSearchOptions{
		PageSize: 20,
		OrderBy:  "id ASC",
		TplName:  USERS,
	})
}

func NewUser(c *context.Context) {
	c.Data["Title"] = c.Tr("admin.users.new_account")
	c.Data["PageIsAdmin"] = true
	c.Data["PageIsAdminUsers"] = true

	c.Success(USER_NEW)
}

func NewUserPost(c *context.Context, f form.AdminCreateUser) {
	c.Data["Title"] = c.Tr("admin.users.new_account")
	c.Data["PageIsAdmin"] = true
	c.Data["PageIsAdminUsers"] = true

	if c.HasError() {
		c.Success(USER_NEW)
		return
	}

	u := &db.User{
		Name:     f.UserName,
		Password: f.Password,
		IsActive: true,
	}

	if err := db.CreateUser(u); err != nil {
		c.Error(err, "create user")
		return
	}
	log.Debugf("Account created by admin %s %s", c.User.Name, u.Name)

	c.Flash.Success(c.Tr("admin.users.new_success", u.Name))
	c.Redirect(conf.Web.Subpath + "/admin/users")
}

func prepareUserInfo(c *context.Context) *db.User {
	u, err := db.UserGetById(c.ParamsInt64(":userid"))
	if err != nil {
		c.Error(err, "get user by ID")
		return nil
	}
	c.Data["User"] = &u

	return &u
}

func EditUser(c *context.Context) {
	c.Data["Title"] = c.Tr("admin.users.edit_account")
	c.Data["PageIsAdmin"] = true
	c.Data["PageIsAdminUsers"] = true

	prepareUserInfo(c)
	if c.Written() {
		return
	}

	c.Success(USER_EDIT)
}

func EditUserPost(c *context.Context, f form.AdminEditUser) {
	c.Data["Title"] = c.Tr("admin.users.edit_account")
	c.Data["PageIsAdmin"] = true
	c.Data["PageIsAdminUsers"] = true

	// u := prepareUserInfo(c)
	// if c.Written() {
	// 	return
	// }

	// if c.HasError() {
	// 	c.Success(USER_EDIT)
	// 	return
	// }

	// fields := strings.Split(f.LoginType, "-")
	// if len(fields) == 2 {
	// 	loginSource := com.StrTo(fields[1]).MustInt64()

	// 	if u.LoginSource != loginSource {
	// 		u.LoginSource = loginSource
	// 	}
	// }

	// if len(f.Password) > 0 {
	// 	u.Passwd = f.Password
	// 	var err error
	// 	if u.Salt, err = db.GetUserSalt(); err != nil {
	// 		c.Error(err, "get user salt")
	// 		return
	// 	}
	// 	u.EncodePassword()
	// }

	// u.LoginName = f.LoginName
	// u.FullName = f.FullName
	// u.Email = f.Email
	// u.Website = f.Website
	// u.Location = f.Location
	// u.MaxRepoCreation = f.MaxRepoCreation
	// u.IsActive = f.Active
	// u.IsAdmin = f.Admin
	// u.AllowGitHook = f.AllowGitHook
	// u.AllowImportLocal = f.AllowImportLocal
	// u.ProhibitLogin = f.ProhibitLogin

	// if err := db.UpdateUser(u); err != nil {
	// 	if db.IsErrEmailAlreadyUsed(err) {
	// 		c.Data["Err_Email"] = true
	// 		c.RenderWithErr(c.Tr("form.email_been_used"), USER_EDIT, &f)
	// 	} else {
	// 		c.Error(err, "update user")
	// 	}
	// 	return
	// }
	// log.Trace("Account profile updated by admin (%s): %s", c.User.Name, u.Name)

	// c.Flash.Success(c.Tr("admin.users.update_profile_success"))
	// c.Redirect(conf.Server.Subpath + "/admin/users/" + c.Params(":userid"))
}
