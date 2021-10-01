package conf

import (
	"net/url"
	"os"
)

// CustomConf returns the absolute path of custom configuration file that is used.
var CustomConf string

// Build time and commit information.
//
// ⚠️ WARNING: should only be set by "-ldflags".
var (
	BuildTime   string
	BuildCommit string
)

var (
	App struct {
		// ⚠️ WARNING: Should only be set by the main package (i.e. "imail.go").
		Version string `ini:"-"`

		Name      string
		BrandName string
		RunUser   string
		RunMode   string
	}

	// log
	Log struct {
		Format   string
		RootPath string
	}

	// mail
	Mail struct {
		Domain      string
		AppDataPath string
	}

	// web settings
	Web struct {
		Enable                   bool
		Port                     int
		Domain                   string
		AppDataPath              string
		AccessControlAllowOrigin string
	}

	// Session settings
	Session struct {
		Provider       string
		ProviderConfig string
		CookieName     string
		CookieSecure   bool
		GCInterval     int64 `ini:"gc_interval"`
		MaxLifeTime    int64
		CSRFCookieName string `ini:"csrf_cookie_name"`
	}

	// Smtp settings
	Smtp struct {
		Port      int
		Enable    bool
		Debug     bool
		SslEnable bool
		SslPort   int
		ModeIn    bool
	}

	// Pop settings
	Pop3 struct {
		Port      int
		Enable    bool
		Debug     bool
		SslEnable bool
		SslPort   int
		ModeIn    bool
	}

	// Imap settings
	Imap struct {
		Port      int
		Enable    bool
		Debug     bool
		SslEnable bool
		SslPort   int
		ModeIn    bool
	}

	//rspamd
	Rspamd struct {
		Enable                bool
		Domain                string
		Password              string
		RecjectConditionScore float64
	}

	//Hook
	Hook struct {
		Enable        bool
		ReceiveScript string
		SendScript    string
	}

	// Security settings
	Security struct {
		InstallLock bool
		SecretKey   string
	}
)

type ServerOpts struct {
	ExternalURL          string `ini:"EXTERNAL_URL"`
	Domain               string
	Protocol             string
	HTTPAddr             string `ini:"HTTP_ADDR"`
	HTTPPort             string `ini:"HTTP_PORT"`
	CertFile             string
	KeyFile              string
	TLSMinVersion        string `ini:"TLS_MIN_VERSION"`
	UnixSocketPermission string
	LocalRootURL         string `ini:"LOCAL_ROOT_URL"`

	OfflineMode      bool
	DisableRouterLog bool
	EnableGzip       bool

	AppDataPath        string
	LoadAssetsFromDisk bool

	LandingURL string `ini:"LANDING_URL"`

	// Derived from other static values
	URL            *url.URL    `ini:"-"` // Parsed URL object of ExternalURL.
	Subpath        string      `ini:"-"` // Subpath found the ExternalURL. Should be empty when not found.
	SubpathDepth   int         `ini:"-"` // The number of slashes found in the Subpath.
	UnixSocketMode os.FileMode `ini:"-"` // Parsed file mode of UnixSocketPermission.
}

// Server settings
var Server ServerOpts

type DatabaseOpts struct {
	Type         string
	Host         string
	Name         string
	User         string
	Password     string
	SSLMode      string `ini:"ssl_mode"`
	Path         string
	Charset      string
	MaxOpenConns int
	MaxIdleConns int
}

// Database settings
var Database DatabaseOpts

type i18nConf struct {
	Langs     []string          `delim:","`
	Names     []string          `delim:","`
	dateLangs map[string]string `ini:"-"`
}

// DateLang transforms standard language locale name to corresponding value in datetime plugin.
func (c *i18nConf) DateLang(lang string) string {
	name, ok := c.dateLangs[lang]
	if ok {
		return name
	}
	return "en"
}

// I18n settings
var I18n *i18nConf
