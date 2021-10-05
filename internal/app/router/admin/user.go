package admin

import (
	// "strings"

	"github.com/midoks/imail/internal/app/context"
	"github.com/midoks/imail/internal/app/form"
	// "github.com/midoks/imail/internal/conf"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/tools/paginater"
	// "github.com/midoks/imail/internal/log"
)

const (
	USERS     = "admin/user/list"
	USER_NEW  = "admin/user/new"
	USER_EDIT = "admin/user/edit"
)

type UserSearchOptions struct {
	Type int
	// Counter  func() int64
	// Ranger   func(int, int) ([]*db.User, error)
	PageSize int
	OrderBy  string
	TplName  string
}

func RenderUserSearch(c *context.Context, opts *UserSearchOptions) {
	page := c.QueryInt("page")
	if page <= 1 {
		page = 1
	}

	var (
		users []*db.User
		count int64
		err   error
	)

	keyword := c.Query("q")
	if len(keyword) == 0 {
		// users, err = opts.Ranger(page, opts.PageSize)
		// if err != nil {
		// 	c.Error(err, "ranger")
		// 	return
		// }
		// count = opts.Counter()
	} else {
		// users, count, err = db.SearchUserByName(&db.SearchUserOptions{
		// 	Keyword:  keyword,
		// 	Type:     opts.Type,
		// 	OrderBy:  opts.OrderBy,
		// 	Page:     page,
		// 	PageSize: opts.PageSize,
		// })
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

	RenderUserSearch(c, &UserSearchOptions{
		Type: 0,
		// Counter:  db.CountUsers,
		// Ranger:   db.ListUsers,
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

}

// func prepareUserInfo(c *context.Context) *db.User {
// 	u, err := db.GetUserByID(c.ParamsInt64(":userid"))
// 	if err != nil {
// 		c.Error(err, "get user by ID")
// 		return nil
// 	}
// 	c.Data["User"] = u

// 	if u.LoginSource > 0 {
// 		c.Data["LoginSource"], err = db.LoginSources.GetByID(u.LoginSource)
// 		if err != nil {
// 			c.Error(err, "get login source by ID")
// 			return nil
// 		}
// 	} else {
// 		c.Data["LoginSource"] = &db.LoginSource{}
// 	}

// 	sources, err := db.LoginSources.List(db.ListLoginSourceOpts{})
// 	if err != nil {
// 		c.Error(err, "list login sources")
// 		return nil
// 	}
// 	c.Data["Sources"] = sources

// 	return u
// }

// func EditUser(c *context.Context) {
// 	c.Data["Title"] = c.Tr("admin.users.edit_account")
// 	c.Data["PageIsAdmin"] = true
// 	c.Data["PageIsAdminUsers"] = true
// 	c.Data["EnableLocalPathMigration"] = conf.Repository.EnableLocalPathMigration

// 	prepareUserInfo(c)
// 	if c.Written() {
// 		return
// 	}

// 	c.Success(USER_EDIT)
// }

// func EditUserPost(c *context.Context, f form.AdminEditUser) {
// 	c.Data["Title"] = c.Tr("admin.users.edit_account")
// 	c.Data["PageIsAdmin"] = true
// 	c.Data["PageIsAdminUsers"] = true
// 	c.Data["EnableLocalPathMigration"] = conf.Repository.EnableLocalPathMigration

// 	u := prepareUserInfo(c)
// 	if c.Written() {
// 		return
// 	}

// 	if c.HasError() {
// 		c.Success(USER_EDIT)
// 		return
// 	}

// 	fields := strings.Split(f.LoginType, "-")
// 	if len(fields) == 2 {
// 		loginSource := com.StrTo(fields[1]).MustInt64()

// 		if u.LoginSource != loginSource {
// 			u.LoginSource = loginSource
// 		}
// 	}

// 	if len(f.Password) > 0 {
// 		u.Passwd = f.Password
// 		var err error
// 		if u.Salt, err = db.GetUserSalt(); err != nil {
// 			c.Error(err, "get user salt")
// 			return
// 		}
// 		u.EncodePassword()
// 	}

// 	u.LoginName = f.LoginName
// 	u.FullName = f.FullName
// 	u.Email = f.Email
// 	u.Website = f.Website
// 	u.Location = f.Location
// 	u.MaxRepoCreation = f.MaxRepoCreation
// 	u.IsActive = f.Active
// 	u.IsAdmin = f.Admin
// 	u.AllowGitHook = f.AllowGitHook
// 	u.AllowImportLocal = f.AllowImportLocal
// 	u.ProhibitLogin = f.ProhibitLogin

// 	if err := db.UpdateUser(u); err != nil {
// 		if db.IsErrEmailAlreadyUsed(err) {
// 			c.Data["Err_Email"] = true
// 			c.RenderWithErr(c.Tr("form.email_been_used"), USER_EDIT, &f)
// 		} else {
// 			c.Error(err, "update user")
// 		}
// 		return
// 	}
// 	log.Trace("Account profile updated by admin (%s): %s", c.User.Name, u.Name)

// 	c.Flash.Success(c.Tr("admin.users.update_profile_success"))
// 	c.Redirect(conf.Server.Subpath + "/admin/users/" + c.Params(":userid"))
// }

// func DeleteUser(c *context.Context) {
// 	u, err := db.GetUserByID(c.ParamsInt64(":userid"))
// 	if err != nil {
// 		c.Error(err, "get user by ID")
// 		return
// 	}

// 	if err = db.DeleteUser(u); err != nil {
// 		switch {
// 		case db.IsErrUserOwnRepos(err):
// 			c.Flash.Error(c.Tr("admin.users.still_own_repo"))
// 			c.JSONSuccess(map[string]interface{}{
// 				"redirect": conf.Server.Subpath + "/admin/users/" + c.Params(":userid"),
// 			})
// 		case db.IsErrUserHasOrgs(err):
// 			c.Flash.Error(c.Tr("admin.users.still_has_org"))
// 			c.JSONSuccess(map[string]interface{}{
// 				"redirect": conf.Server.Subpath + "/admin/users/" + c.Params(":userid"),
// 			})
// 		default:
// 			c.Error(err, "delete user")
// 		}
// 		return
// 	}
// 	log.Trace("Account deleted by admin (%s): %s", c.User.Name, u.Name)

// 	c.Flash.Success(c.Tr("admin.users.deletion_success"))
// 	c.JSONSuccess(map[string]interface{}{
// 		"redirect": conf.Server.Subpath + "/admin/users",
// 	})
// }
