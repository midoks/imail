package app

import (
	"fmt"
	"path/filepath"

	"github.com/go-macaron/csrf"
	"github.com/go-macaron/gzip"
	"github.com/go-macaron/i18n"
	"github.com/go-macaron/session"
	"github.com/midoks/imail/internal/app/context"
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

	fmt.Println(conf.I18n.Langs)
	fmt.Println(conf.I18n.Names)

	m.Use(i18n.I18n(i18n.Options{
		Directory:       filepath.Join(conf.WorkDir(), "conf", "locale"),
		CustomDirectory: filepath.Join(conf.CustomDir(), "conf", "locale"),
		Langs:           conf.I18n.Langs,
		Names:           conf.I18n.Names,
		Format:          "locale_%s.ini",
		DefaultLang:     "en-US",
		Redirect:        true,
	}))

	return m
}

func setRouter(m *macaron.Macaron) *macaron.Macaron {

	// if !conf.Security.InstallLock {
	// 	c.RedirectSubpath("/install")
	// 	return
	// }

	reqSignIn := context.Toggle(&context.ToggleOptions{SignInRequired: true})
	// ignSignIn := context.Toggle(&context.ToggleOptions{SignInRequired: conf.Auth.RequireSigninView})
	reqSignOut := context.Toggle(&context.ToggleOptions{SignOutRequired: true})

	m.SetAutoHead(true)

	m.Group("", func() {
		m.Get("/", reqSignIn, func(ctx *context.Context) {
			ctx.Success("home")
		})

		m.Get("/install", func(ctx *context.Context) {
			ctx.HTML(200, "install")
		}, reqSignOut)
	}, session.Sessioner(session.Options{
		Provider:       conf.Session.Provider,
		ProviderConfig: conf.Session.ProviderConfig,
		CookieName:     conf.Session.CookieName,
		CookiePath:     conf.Server.Subpath,
		Gclifetime:     conf.Session.GCInterval,
		Maxlifetime:    conf.Session.MaxLifeTime,
		Secure:         conf.Session.CookieSecure,
	}), csrf.Csrfer(csrf.Options{
		Secret:         conf.Security.SecretKey,
		Header:         "X-CSRF-Token",
		Cookie:         conf.Session.CSRFCookieName,
		CookieDomain:   conf.Server.URL.Hostname(),
		CookiePath:     conf.Server.Subpath,
		CookieHttpOnly: true,
		SetCookie:      true,
		Secure:         conf.Server.URL.Scheme == "https",
	}), context.Contexter())
	return m
}

func Start(port int) {
	m := newMacaron()
	m = setRouter(m)
	fmt.Println(port)
	m.Run(port)
}
