package app

import (
	"fmt"

	"github.com/go-macaron/csrf"
	"github.com/go-macaron/gzip"
	"github.com/go-macaron/session"
	"github.com/midoks/imail/internal/app/context"
	"github.com/midoks/imail/internal/app/template"
	"github.com/midoks/imail/internal/conf"
	// "github.com/midoks/imail/internal/log"
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
		Directory: "templates/default",
		Funcs:     template.FuncMap(),
	})
	m.Use(opt)

	// localeNames, err := conf.AssetDir("conf/locale")
	// if err != nil {
	// 	log.Fatal("Failed to list locale files: %v", err)
	// }
	// localeFiles := make(map[string][]byte)
	// for _, name := range localeNames {
	// 	localeFiles[name] = conf.MustAsset("conf/locale/" + name)
	// }
	// m.Use(i18n.I18n(i18n.Options{
	// 	SubURL:          conf.Server.Subpath,
	// 	Files:           localeFiles,
	// 	CustomDirectory: filepath.Join(conf.CustomDir(), "conf", "locale"),
	// 	Langs:           conf.I18n.Langs,
	// 	Names:           conf.I18n.Names,
	// 	DefaultLang:     "en-US",
	// 	Redirect:        true,
	// }))

	m.SetAutoHead(true)
	return m
}

func setRouter(m *macaron.Macaron) *macaron.Macaron {

	// if !conf.Security.InstallLock {
	// 	c.RedirectSubpath("/install")
	// 	return
	// }

	m.Group("", func() {
		m.Get("/", func(ctx *context.Context) {
			ctx.Success("home")
		})

		m.Get("/install", func(ctx *macaron.Context) {
			ctx.HTML(200, "install")
		})
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
