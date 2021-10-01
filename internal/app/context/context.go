// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package context

import (
	"net/http"
	"time"

	"github.com/go-macaron/cache"
	"github.com/go-macaron/csrf"
	"gopkg.in/macaron.v1"
	// "github.com/go-macaron/i18n"
	"github.com/go-macaron/session"

	"github.com/midoks/imail/internal/app/template"
	"github.com/midoks/imail/internal/conf"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/log"
)

// Context represents context of a request.
type Context struct {
	*macaron.Context
	Cache   cache.Cache
	csrf    csrf.CSRF
	Flash   *session.Flash
	Session session.Store

	Link        string // Current request URL
	User        *db.User
	IsLogged    bool
	IsBasicAuth bool
	IsTokenAuth bool
}

// RawTitle sets the "Title" field in template data.
func (c *Context) RawTitle(title string) {
	c.Data["Title"] = title
}

// PageIs sets "PageIsxxx" field in template data.
func (c *Context) PageIs(name string) {
	c.Data["PageIs"+name] = true
}

// Require sets "Requirexxx" field in template data.
func (c *Context) Require(name string) {
	c.Data["Require"+name] = true
}

// FormErr sets "Err_xxx" field in template data.
func (c *Context) FormErr(names ...string) {
	for i := range names {
		c.Data["Err_"+names[i]] = true
	}
}

func (c *Context) GetErrMsg() string {
	return c.Data["ErrorMsg"].(string)
}

// HasValue returns true if value of given name exists.
func (c *Context) HasValue(name string) bool {
	_, ok := c.Data[name]
	return ok
}

// HTML responses template with given status.
func (c *Context) HTML(status int, name string) {
	log.Trace("Template: %s", name)
	c.Context.HTML(status, name)
}

// Success responses template with status http.StatusOK.
func (c *Context) Success(name string) {
	c.HTML(http.StatusOK, name)
}

// RedirectSubpath responses redirection with given location and status.
// It prepends setting.Server.Subpath to the location string.
func (c *Context) RedirectSubpath(location string, status ...int) {
	c.Redirect(conf.Server.Subpath+location, status...)
}

// Contexter initializes a classic context for a request.
//l i18n.Locale, cache cache.Cache, sess session.Store, f *session.Flash,
func Contexter() macaron.Handler {
	return func(ctx *macaron.Context, sess session.Store, f *session.Flash, x csrf.CSRF) {
		c := &Context{
			Context: ctx,
			// Cache:   cache,
			csrf:    x,
			Flash:   f,
			Session: sess,
		}

		c.Data["PageStartTime"] = time.Now()

		if len(conf.Web.AccessControlAllowOrigin) > 0 {
			c.Header().Set("Access-Control-Allow-Origin", conf.Web.AccessControlAllowOrigin)
			c.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Header().Set("Access-Control-Max-Age", "3600")
			c.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
		}

		// Get user from session or header when possible
		// c.User, c.IsBasicAuth, c.IsTokenAuth = authenticatedUser(c.Context, c.Session)

		// if c.User != nil {
		// 	c.IsLogged = true
		// 	c.Data["IsLogged"] = c.IsLogged
		// 	c.Data["LoggedUser"] = c.User
		// 	c.Data["LoggedUserID"] = c.User.ID
		// 	c.Data["LoggedUserName"] = c.User.Name
		// 	c.Data["IsAdmin"] = c.User.IsAdmin
		// } else {
		// 	c.Data["LoggedUserID"] = 0
		// 	c.Data["LoggedUserName"] = ""
		// }

		c.Data["CSRFToken"] = x.GetToken()
		c.Data["CSRFTokenHTML"] = template.Safe(`<input type="hidden" name="_csrf" value="` + x.GetToken() + `">`)
		// log.Trace("Session ID: %s", sess.ID())
		log.Trace("CSRF Token: %v", c.Data["CSRFToken"])

		// c.Data["ShowRegistrationButton"] = !conf.Auth.DisableRegistration
		// c.Data["ShowFooterBranding"] = conf.Other.ShowFooterBranding

		// c.renderNoticeBanner()

		c.Header().Set("X-Content-Type-Options", "nosniff")
		c.Header().Set("X-Frame-Options", "DENY")

		ctx.Map(c)
	}
}
