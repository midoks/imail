package conf

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"github.com/midoks/imail/internal/log"
	"github.com/midoks/imail/internal/tools"
	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
)

var confToml *toml.Tree
var err error

var IsLoadedVar bool

func Load(path string) error {

	confToml, err = toml.LoadFile(path) //load config file

	if err != nil {
		panic("config init error")
		IsLoadedVar = false
		return err
	}
	IsLoadedVar = true
	http_enable, err := GetBool("http.enable", false)
	if err == nil && http_enable {
		ipWhiteList := strings.Split(GetString("http.ip_white", "*"), ",")
		if InSliceString("*", ipWhiteList) {
			return nil
		}
		for _, ip := range ipWhiteList {
			if strings.Contains(ip, "/") {
				_, _, err = net.ParseCIDR(ip)
				if err != nil {
					return err
				}
			} else {
				if net.ParseIP(ip) == nil {
					return errors.New(fmt.Sprint(ip, ", Invalid whitelist"))
				}
			}

		}

	}
	return nil
}

func LoadString(content string) error {
	confToml, err = toml.Load(content) //load config string
	return err
}

func InSliceString(v string, sl []string) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}

func IsLoaded() bool {
	return IsLoadedVar
}

func GetString(key string, def string) string {
	v := confToml.Get(key)

	if reflect.TypeOf(v) == nil {
		return def
	}

	return v.(string)
}

func GetInt64(key string, def int64) (int64, error) {
	v := confToml.Get(key)

	if reflect.TypeOf(v) == nil {
		return def, nil
	}

	if reflect.TypeOf(v).String() != "int64" {
		return def, errors.New(key + " type is error, expect is int64 type!")
	}
	return v.(int64), nil
}

func GetInt(key string, def int) (int, error) {
	v := confToml.Get(key)

	if reflect.TypeOf(v) == nil {
		return def, nil
	}

	if reflect.TypeOf(v).String() != "int64" {
		return def, errors.New(key + " type is error, expect is int type!")
	}

	vv := v.(int64)

	vInt := *(*int)(unsafe.Pointer(&vv))
	return vInt, nil
}

func GetFloat64(key string, def float64) (float64, error) {
	v := confToml.Get(key)

	if reflect.TypeOf(v) == nil {
		return def, nil
	}

	if reflect.TypeOf(v).String() != "float64" {
		return def, errors.New(key + " type is error, expect is float64 type!")
	}
	return v.(float64), nil
}

func GetBool(key string, def bool) (bool, error) {
	v := confToml.Get(key)

	if reflect.TypeOf(v) == nil {
		return def, nil
	}

	if reflect.TypeOf(v).String() != "bool" {
		return def, errors.New(key + " type is error, expect is bool type!")
	}

	return v.(bool), nil
}

// File is the configuration object.
var File *ini.File

func ReadFile(file string) (string, error) {
	f, err := os.OpenFile(file, os.O_RDONLY, 0600)
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	return string(b), err
}

func Init(customConf string) error {

	definedConf, _ := ReadFile("conf/app.defined.conf")

	File, err = ini.LoadSources(ini.LoadOptions{
		IgnoreInlineComment: true,
	}, []byte(definedConf))
	if err != nil {
		return errors.Wrap(err, "parse 'conf/app.conf'")
	}

	File.NameMapper = ini.SnackCase

	if customConf == "" {
		customConf = filepath.Join(CustomDir(), "conf", "app.conf")
	} else {
		customConf, err = filepath.Abs(customConf)
		if err != nil {
			return errors.Wrap(err, "get absolute path")
		}
	}
	CustomConf = customConf

	if tools.IsFile(customConf) {
		if err = File.Append(customConf); err != nil {
			return errors.Wrapf(err, "append %q", customConf)
		}
	} else {
		log.Warnf("Custom config %q not found. Ignore this warning if you're running for the first time", customConf)
	}

	if err = File.Section(ini.DefaultSection).MapTo(&App); err != nil {
		return errors.Wrap(err, "mapping default section")
	}

	// ***************************
	// ----- Server settings -----
	// ***************************

	if err = File.Section("server").MapTo(&Server); err != nil {
		return errors.Wrap(err, "mapping [server] section")
	}
	Server.AppDataPath = ensureAbs(Server.AppDataPath)

	if !strings.HasSuffix(Server.ExternalURL, "/") {
		Server.ExternalURL += "/"
	}
	Server.URL, err = url.Parse(Server.ExternalURL)
	if err != nil {
		return errors.Wrapf(err, "parse '[server] EXTERNAL_URL' %q", err)
	}

	// Subpath should start with '/' and end without '/', i.e. '/{subpath}'.
	Server.Subpath = strings.TrimRight(Server.URL.Path, "/")
	Server.SubpathDepth = strings.Count(Server.Subpath, "/")

	unixSocketMode, err := strconv.ParseUint(Server.UnixSocketPermission, 8, 32)
	if err != nil {
		return errors.Wrapf(err, "parse '[server] UNIX_SOCKET_PERMISSION' %q", Server.UnixSocketPermission)
	}
	if unixSocketMode > 0777 {
		unixSocketMode = 0666
	}
	Server.UnixSocketMode = os.FileMode(unixSocketMode)

	// ***************************
	// ----- SMTP settings -----
	// ***************************
	if err = File.Section("smtp").MapTo(&Smtp); err != nil {
		return errors.Wrap(err, "mapping [smtp] section")
	}

	// ***************************
	// ----- Pop settings -----
	// ***************************
	if err = File.Section("pop").MapTo(&Pop); err != nil {
		return errors.Wrap(err, "mapping [pop] section")
	}

	// ***************************
	// ----- Imap settings -----
	// ***************************
	if err = File.Section("imap").MapTo(&Imap); err != nil {
		return errors.Wrap(err, "mapping [imap] section")
	}

	// ****************************
	// ----- Session settings -----
	// ****************************

	if err = File.Section("session").MapTo(&Session); err != nil {
		return errors.Wrap(err, "mapping [session] section")
	}

	// *****************************
	// ----- Security settings -----
	// *****************************

	if err = File.Section("security").MapTo(&Security); err != nil {
		return errors.Wrap(err, "mapping [security] section")
	}

	// Check run user when the install is locked.
	if Security.InstallLock {
		currentUser, match := CheckRunUser(App.RunUser)
		if !match {
			return fmt.Errorf("user configured to run imail is %q, but the current user is %q", App.RunUser, currentUser)
		}
	}

	return nil
}
