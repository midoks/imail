package conf

import (
	"net/url"
	"os"
)

// CustomConf returns the absolute path of custom configuration file that is used.
var CustomConf string

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
		Format string
	}

	// mail
	Mail struct {
		Domain      string
		AppDataPath string
	}

	// web settings
	Web struct {
		Port                     int
		Enable                   bool
		AccessControlAllowOrigin string
	}

	// Session settings
	Session struct {
		Provider       string
		ProviderConfig string
		CookieName     string
		CookieSecure   bool
		GCInterval     int64 `ini:"GC_INTERVAL"`
		MaxLifeTime    int64
		CSRFCookieName string `ini:"CSRF_COOKIE_NAME"`
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
	SSLMode      string `ini:"SSL_MODE"`
	Path         string
	Charset      string
	MaxOpenConns int
	MaxIdleConns int
}

// Database settings
var Database DatabaseOpts
