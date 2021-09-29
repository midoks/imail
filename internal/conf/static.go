package conf

import ()

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
)
