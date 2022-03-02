// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package context

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-macaron/cache"
	"github.com/go-macaron/csrf"
	"github.com/go-macaron/i18n"
	"github.com/go-macaron/session"
	"gopkg.in/macaron.v1"

	"github.com/midoks/imail/internal/app/form"
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

//json api common data
type JsonMsg struct {
	Code int64       `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// RawTitle sets the "Title" field in template data.
func (c *Context) RawTitle(title string) {
	c.Data["Title"] = title
}

// Title localizes the "Title" field in template data.
func (c *Context) Title(locale string) {
	c.RawTitle(c.Tr(locale))
}

// RenderWithErr used for page has form validation but need to prompt error to users.
func (c *Context) RenderWithErr(msg, tpl string, f interface{}) {
	if f != nil {
		form.Assign(f, c.Data)
	}
	c.Flash.ErrorMsg = msg
	c.Data["Flash"] = c.Flash
	c.HTML(http.StatusOK, tpl)
}

// HasError returns true if error occurs in form validation.
func (c *Context) HasError() bool {
	hasErr, ok := c.Data["HasError"]
	if !ok {
		return false
	}
	c.Flash.ErrorMsg = c.Data["ErrorMsg"].(string)
	c.Data["Flash"] = c.Flash
	return hasErr.(bool)
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
	log.Infof("Template:%s", name)
	c.Context.HTML(status, name)
}

// Success responses template with status http.StatusOK.
func (c *Context) Success(name string) {
	c.HTML(http.StatusOK, name)
}

// JSONSuccess responses JSON with status http.StatusOK.
func (c *Context) JSONSuccess(data interface{}) {
	c.JSON(http.StatusOK, data)
}

//JSON Success Message
func (c *Context) OK(msg string) {
	c.JSONSuccess(JsonMsg{Code: 0, Msg: msg})
}

func (c *Context) OKDATA(msg string, data interface{}) {
	c.JSONSuccess(JsonMsg{Code: 0, Msg: msg, Data: data})
}

//JSON Fail Message
func (c *Context) Fail(code int64, msg string) {
	c.JSONSuccess(JsonMsg{Code: code, Msg: msg})
}

// NotFound renders the 404 page.
func (c *Context) NotFound() {
	c.Title("status.page_not_found")
	c.HTML(http.StatusNotFound, fmt.Sprintf("status/%d", http.StatusNotFound))
}

// Error renders the 500 page.
func (c *Context) Error(err error, msg string) {
	// log.ErrorDepth(4, "%s: %v", msg, err)
	c.Title("status.internal_server_error")

	// Only in non-production mode or admin can see the actual error message.
	if !conf.IsProdMode() || (c.IsLogged && c.User.IsAdmin) {
		c.Data["ErrorMsg"] = err
	}
	c.HTML(http.StatusInternalServerError, fmt.Sprintf("status/%d", http.StatusInternalServerError))
}

// Errorf renders the 500 response with formatted message.
func (c *Context) Errorf(err error, format string, args ...interface{}) {
	c.Error(err, fmt.Sprintf(format, args...))
}

// NotFoundOrError responses with 404 page for not found error and 500 page otherwise.
func (c *Context) NotFoundOrError(err error, msg string) {
	// if errutil.IsNotFound(err) {
	// 	c.NotFound()
	// 	return
	// }
	c.Error(err, msg)
}

// RedirectSubpath responses redirection with given location and status.
// It prepends setting.Server.Subpath to the location string.
func (c *Context) RedirectSubpath(location string, status ...int) {
	c.Redirect(conf.Web.Subpath+location, status...)
}

// Contexter initializes a classic context for a request.
func Contexter() macaron.Handler {
	return func(ctx *macaron.Context, l i18n.Locale, cache cache.Cache, sess session.Store, f *session.Flash, x csrf.CSRF) {
		c := &Context{
			Context: ctx,
			Cache:   cache,
			csrf:    x,
			Flash:   f,
			Session: sess,
		}

		c.Data["NowLang"] = l.Lang
		c.Data["PageStartTime"] = time.Now()

		if len(conf.Web.AccessControlAllowOrigin) > 0 {
			c.Header().Set("Access-Control-Allow-Origin", conf.Web.AccessControlAllowOrigin)
			c.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Header().Set("Access-Control-Max-Age", "3600")
			c.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
		}

		// Get user from session or header when possible
		uid := c.Session.Get("uid")
		if uid != nil {
			u, err := db.UserGetById(uid.(int64))
			if err == nil {
				c.IsLogged = true
				c.Data["IsLogged"] = c.IsLogged
				c.Data["LoggedUser"] = u
				c.Data["LoggedUserID"] = u.Id
				c.Data["LoggedUserName"] = u.Name
				c.Data["IsAdmin"] = u.IsAdmin
			}

			c.User = &u
			c.Data["MenuDomains"], _ = db.DomainVaildList(1, 10)
		} else {
			c.Data["LoggedUserID"] = 0
			c.Data["LoggedUserName"] = ""
		}

		c.Data["CSRFToken"] = x.GetToken()
		c.Data["CSRFTokenHTML"] = template.Safe(`<input type="hidden" name="_csrf" value="` + x.GetToken() + `">`)
		// log.Debugf("Session ID: %s", sess.ID())
		// log.Debugf("CSRF Token: %s", c.Data["CSRFToken"])

		c.Data["ShowRegistrationButton"] = !conf.Auth.DisableRegistration
		// c.Data["ShowFooterBranding"] = conf.Other.ShowFooterBranding

		// c.renderNoticeBanner()

		// avoid iframe use by other.
		c.Header().Set("X-Content-Type-Options", "nosniff")
		c.Header().Set("X-Frame-Options", "DENY")

		ctx.Map(c)
	}
}
