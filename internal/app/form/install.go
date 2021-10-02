package form

import ()

type Install struct {
	DbType   string `binding:"Required"`
	DbHost   string
	DbUser   string
	DbPasswd string
	DbName   string
	SslMode  string
	DbPath   string

	AppName           string `binding:"Required" locale:"install.app_name"`
	RepoRootPath      string `binding:"Required"`
	RunUser           string `binding:"Required"`
	Domain            string `binding:"Required"`
	HttpPort          string `binding:"Required"`
	LogRootPath       string `binding:"Required"`
	EnableConsoleMode bool

	OfflineMode           bool
	DisableGravatar       bool
	EnableFederatedAvatar bool
	DisableRegistration   bool
	EnableCaptcha         bool
	RequireSignInView     bool

	AdminName          string `binding:"OmitEmpty;AlphaDashDot;MaxSize(30)" locale:"install.admin_name"`
	AdminPasswd        string `binding:"OmitEmpty;MaxSize(255)" locale:"install.admin_password"`
	AdminConfirmPasswd string
	AdminEmail         string `binding:"OmitEmpty;MinSize(3);MaxSize(254);Include(@)" locale:"install.admin_email"`
}
