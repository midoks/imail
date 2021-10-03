package app

import (
	"fmt"
	"path/filepath"

	"github.com/go-macaron/binding"
	"github.com/go-macaron/cache"
	"github.com/go-macaron/captcha"
	"github.com/go-macaron/csrf"
	"github.com/go-macaron/gzip"
	"github.com/go-macaron/i18n"
	"github.com/go-macaron/session"

	"github.com/midoks/imail/internal/app/context"
	"github.com/midoks/imail/internal/app/form"
	"github.com/midoks/imail/internal/app/router"
	"github.com/midoks/imail/internal/app/router/admin"
	"github.com/midoks/imail/internal/app/router/user"
	"github.com/midoks/imail/internal/app/template"
	"github.com/midoks/imail/internal/conf"
	// "github.com/midoks/imail/internal/log"
	// "github.com/midoks/imail/internal/tools"
	"gopkg.in/macaron.v1"
)

func newMacaron() *macaron.Macaron {
	m := macaron.New()
	m.Use(macaron.Logger())
	m.Use(gzip.Gziper())
	m.Use(macaron.Logger())
	m.Use(macaron.Recovery())

	m.Use(macaron.Static("public"))

	opt := macaron.Renderer(macaron.RenderOptions{
		Directory: "templates",
		Funcs:     template.FuncMap(),
	})
	m.Use(opt)

	m.Use(i18n.I18n(i18n.Options{
		Directory:       filepath.Join(conf.WorkDir(), "conf", "locale"),
		CustomDirectory: filepath.Join(conf.CustomDir(), "conf", "locale"),
		Langs:           conf.I18n.Langs,
		Names:           conf.I18n.Names,
		Format:          "locale_%s.ini",
		DefaultLang:     "en-US",
		Redirect:        true,
	}))

	m.Use(cache.Cacher(cache.Options{
		Adapter:       conf.Cache.Adapter,
		AdapterConfig: conf.Cache.Host,
		Interval:      conf.Cache.Interval,
	}))

	m.Use(captcha.Captchaer(captcha.Options{
		SubURL: conf.Web.Subpath,
	}))

	return m
}

func setRouter(m *macaron.Macaron) *macaron.Macaron {

	reqSignIn := context.Toggle(&context.ToggleOptions{SignInRequired: true})
	// ignSignIn := context.Toggle(&context.ToggleOptions{SignInRequired: conf.Auth.RequireSigninView})
	reqSignOut := context.Toggle(&context.ToggleOptions{SignOutRequired: true})

	bindIgnErr := binding.BindIgnErr

	m.SetAutoHead(true)

	m.Group("", func() {
		m.Get("/", reqSignIn, router.Home)

		m.Group("/user", func() {
			m.Group("/login", func() {
				m.Combo("").Get(user.Login).Post(bindIgnErr(form.SignIn{}), user.LoginPost)
			})

			m.Get("/sign_up", user.SignUp)
			m.Post("/sign_up", bindIgnErr(form.Register{}), user.SignUpPost)
		}, reqSignOut)

		m.Combo("/install", router.InstallInit).Get(router.Install).Post(bindIgnErr(form.Install{}), router.InstallPost)

		reqAdmin := context.Toggle(&context.ToggleOptions{SignInRequired: true, AdminRequired: true})

		// ***** START: Admin *****
		m.Group("/admin", func() {
			m.Combo("").Get(admin.Dashboard) //.Post(admin.Operation) // "/admin"
			m.Get("/config", admin.Config)
			// m.Post("/config/test_mail", admin.SendTestMail)
			m.Get("/monitor", admin.Monitor)

			// m.Group("/users", func() {
			// 	m.Get("", admin.Users)
			// 	m.Combo("/new").Get(admin.NewUser).Post(bindIgnErr(form.AdminCrateUser{}), admin.NewUserPost)
			// 	m.Combo("/:userid").Get(admin.EditUser).Post(bindIgnErr(form.AdminEditUser{}), admin.EditUserPost)
			// 	m.Post("/:userid/delete", admin.DeleteUser)
			// })

		}, reqAdmin)
		// ***** END: Admin *****

	}, session.Sessioner(session.Options{
		Provider:       conf.Session.Provider,
		ProviderConfig: conf.Session.ProviderConfig,
		CookieName:     conf.Session.CookieName,
		CookiePath:     conf.Web.Subpath,
		Gclifetime:     conf.Session.GCInterval,
		Maxlifetime:    conf.Session.MaxLifeTime,
		Secure:         conf.Session.CookieSecure,
	}), csrf.Csrfer(csrf.Options{
		Secret:         conf.Security.SecretKey,
		Header:         "X-CSRF-Token",
		Cookie:         conf.Session.CSRFCookieName,
		CookieDomain:   conf.Web.URL.Hostname(),
		CookiePath:     conf.Web.Subpath,
		CookieHttpOnly: true,
		SetCookie:      true,
		Secure:         conf.Web.URL.Scheme == "https",
	}), context.Contexter())
	return m
}

func Start(port string) {
	m := newMacaron()
	m = setRouter(m)
	fmt.Println(port)
	m.Run(1080)
}
