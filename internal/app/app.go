package app

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	"gopkg.in/macaron.v1"

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
	"github.com/midoks/imail/internal/app/router/mail"
	"github.com/midoks/imail/internal/app/router/user"
	"github.com/midoks/imail/internal/app/template"
	"github.com/midoks/imail/internal/assets/public"
	"github.com/midoks/imail/internal/assets/templates"
	"github.com/midoks/imail/internal/conf"
)

func newMacaron() *macaron.Macaron {
	m := macaron.New()
	m.Use(macaron.Logger())
	m.Use(gzip.Gziper())
	m.Use(macaron.Logger())
	m.Use(macaron.Recovery())

	m.Use(macaron.Static(
		filepath.Join(conf.CustomDir(), "public"),
		macaron.StaticOptions{
			SkipLogging: true,
		},
	))

	var publicFs http.FileSystem
	if !conf.Web.LoadAssetsFromDisk {
		publicFs = public.NewFileSystem()
	}
	m.Use(macaron.Static(
		filepath.Join(conf.WorkDir(), "public"),
		macaron.StaticOptions{
			FileSystem:  publicFs,
			SkipLogging: false,
		},
	))

	//template start
	renderOpt := macaron.RenderOptions{
		Directory:         filepath.Join(conf.WorkDir(), "templates"),
		AppendDirectories: []string{filepath.Join(conf.CustomDir(), "templates")},
		Funcs:             template.FuncMap(),
		IndentJSON:        macaron.Env != macaron.PROD,
	}
	if !conf.Web.LoadAssetsFromDisk {
		renderOpt.TemplateFileSystem = templates.NewTemplateFileSystem("", renderOpt.AppendDirectories[0])
	}

	m.Use(macaron.Renderer(renderOpt))
	//template end

	localeNames, _ := conf.AssetDir("conf/locale")
	localeFiles := make(map[string][]byte)
	for _, name := range localeNames {
		localeFiles[name] = conf.MustAsset("conf/locale/" + name)
	}
	m.Use(i18n.I18n(i18n.Options{
		CustomDirectory: filepath.Join(conf.CustomDir(), "conf", "locale"),
		Files:           localeFiles,
		Langs:           conf.I18n.Langs,
		Names:           conf.I18n.Names,
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
		m.Combo("/install", router.InstallInit).Get(router.Install).Post(bindIgnErr(form.Install{}), router.InstallPost)

		m.Get("/", reqSignIn, router.Home)
		m.Group("/user", func() {
			m.Group("/login", func() {
				m.Combo("").Get(user.Login).Post(bindIgnErr(form.SignIn{}), user.LoginPost)
			})

			m.Get("/sign_up", user.SignUp)
			m.Post("/sign_up", bindIgnErr(form.Register{}), user.SignUpPost)
			m.Post("/logout", user.SignOut)
		}, reqSignOut)

		// ***** START: User *****
		m.Group("/user/settings", func() {
			m.Get("", user.Settings)
			m.Post("", bindIgnErr(form.UpdateProfile{}), user.SettingsPost)

			m.Get("/authpassword", user.SettingsAuthPassword)
			m.Post("/authpassword", bindIgnErr(form.Empty{}), user.SettingsAuthPasswordPost)

			m.Get("/password", user.SettingsPassword)
			m.Post("/password", bindIgnErr(form.ChangePassword{}), user.SettingsPasswordPost)

			m.Get("/client", user.SettingsClient)
		}, reqSignIn, func(c *context.Context) {
			c.Data["PageIsUserSettings"] = true
		})
		// ***** END: User *****

		// ***** START: Mail *****
		m.Group("/mail", func() {
			m.Get("", mail.Mail)
			m.Combo("/new").Get(mail.New).Post(bindIgnErr(form.SendMail{}), mail.NewPost)

			m.Combo("/flags").Get(mail.Flags)
			m.Combo("/sent").Get(mail.Sent)
			m.Combo("/deleted").Get(mail.Deleted)
			m.Combo("/draft").Get(mail.Draft)
			m.Combo("/junk").Get(mail.Junk)
			m.Combo("/content/:id").Get(mail.Content)
			m.Combo("/content/:id/html").Get(mail.ContentHtml)
			m.Combo("/content/:id/download").Get(mail.ContentDownload)
			m.Combo("/content/:id/attach/:aid").Get(mail.ContentAttach)
			m.Combo("/demo/:id").Get(mail.ContentDemo)

		}, reqSignIn, func(c *context.Context) {
			c.Data["PageIsMail"] = true
		})

		m.Group("/mailbox/:bid", func() {
			m.Get("", mail.Mail)
			m.Combo("/new").Get(mail.New).Post(bindIgnErr(form.SendMail{}), mail.NewPost)

			m.Combo("/flags").Get(mail.Flags)
			m.Combo("/sent").Get(mail.Sent)
			m.Combo("/deleted").Get(mail.Deleted)
			m.Combo("/junk").Get(mail.Junk)
			m.Combo("/draft").Get(mail.Draft)
			m.Combo("/content/:id").Get(mail.Content)

		}, reqSignIn, func(c *context.Context) {
			c.Data["PageIsMail"] = true
		})
		// ***** END: Mail *****

		reqAdmin := context.Toggle(&context.ToggleOptions{SignInRequired: true, AdminRequired: true})

		// ***** START: Admin *****
		m.Group("/admin", func() {
			m.Combo("").Get(admin.Dashboard) //.Post(admin.Operation) // "/admin"
			m.Get("/config", admin.Config)
			// m.Post("/config/test_mail", admin.SendTestMail)
			m.Get("/monitor", admin.Monitor)

			m.Group("/domain", func() {
				m.Get("", admin.Domain)
				m.Combo("/new").Get(admin.NewDomain).Post(bindIgnErr(form.AdminCreateDomain{}), admin.NewDomainPost)
				m.Combo("/delete/:id").Get(admin.DeleteDomain)
				m.Combo("/check/:id").Get(admin.CheckDomain)
				m.Combo("/default/:id").Get(admin.SetDefaultDomain)
				m.Combo("/info/:domain").Get(admin.InfoDomain)
			})
			m.Group("/users", func() {
				m.Get("", admin.Users)
				m.Combo("/new").Get(admin.NewUser).Post(bindIgnErr(form.AdminCreateUser{}), admin.NewUserPost)
				m.Combo("/:userid").Get(admin.EditUser).Post(bindIgnErr(form.AdminEditUser{}), admin.EditUserPost)
			})

		}, reqAdmin)

		m.Group("/api", func() {
			m.Group("/mail", func() {
				m.Post("/read", bindIgnErr(form.MailIDs{}), mail.ApiRead)
				m.Post("/unread", bindIgnErr(form.MailIDs{}), mail.ApiUnread)

				m.Post("/star", bindIgnErr(form.MailIDs{}), mail.ApiStar)
				m.Post("/unstar", bindIgnErr(form.MailIDs{}), mail.ApiUnStar)
				m.Post("/move", bindIgnErr(form.MailIDs{}), mail.ApiMove)
				m.Post("/deleted", bindIgnErr(form.MailIDs{}), mail.ApiDeleted)
				m.Post("/hard_deleted", bindIgnErr(form.MailIDs{}), mail.ApiHardDeleted)
			})

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

	portInt, err := strconv.Atoi(port)
	if err != nil {
		fmt.Println("port need number!")
	}
	m.Run(portInt)
}
