package conf

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	// "github.com/midoks/imail/internal/log"
	"github.com/midoks/imail/internal/tools"
	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
)

// File is the configuration object.
var File *ini.File

func ReadFile(file string) (string, error) {
	f, err := os.OpenFile(file, os.O_RDONLY, 0600)
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	return string(b), err
}

func Init(customConf string) error {
	fmt.Println("init conf")

	appConf := filepath.Join(WorkDir(), "conf", "app.conf")
	definedConf, _ := ReadFile(appConf)
	File, err := ini.LoadSources(ini.LoadOptions{
		IgnoreInlineComment: true,
	}, []byte(definedConf))
	if err != nil {
		return errors.Wrap(err, "parse 'conf/app.conf'")
	}

	File.NameMapper = ini.TitleUnderscore

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
		// log.Warnf("Custom config %q not found. Ignore this warning if you're running for the first time", customConf)
	}

	if err = File.Section(ini.DefaultSection).MapTo(&App); err != nil {
		return errors.Wrap(err, "mapping default section")
	}

	// ***************************
	// ----- Log settings -----
	// ***************************
	if err = File.Section("log").MapTo(&Log); err != nil {
		return errors.Wrap(err, "mapping [log] section")
	}

	// ***************************
	// ----- Database settings -----
	// ***************************
	if err = File.Section("database").MapTo(&Database); err != nil {
		return errors.Wrap(err, "mapping [log] section")
	}

	// ****************************
	// ----- Web settings -----
	// ****************************

	if err = File.Section("web").MapTo(&Web); err != nil {
		return errors.Wrap(err, "mapping [web] section")
	}

	Web.AppDataPath = ensureAbs(Web.AppDataPath)

	if !strings.HasSuffix(Web.ExternalURL, "/") {
		Web.ExternalURL += "/"
	}
	Web.URL, err = url.Parse(Web.ExternalURL)
	if err != nil {
		return errors.Wrapf(err, "parse '[server] EXTERNAL_URL' %q", err)
	}

	// Subpath should start with '/' and end without '/', i.e. '/{subpath}'.
	Web.Subpath = strings.TrimRight(Web.URL.Path, "/")
	Web.SubpathDepth = strings.Count(Web.Subpath, "/")

	unixSocketMode, err := strconv.ParseUint(Web.UnixSocketPermission, 8, 32)
	if err != nil {
		return errors.Wrapf(err, "parse '[server] unix_socket_permission' %q", Web.UnixSocketPermission)
	}
	if unixSocketMode > 0777 {
		unixSocketMode = 0666
	}
	Web.UnixSocketMode = os.FileMode(unixSocketMode)

	// ****************************
	// ----- Session settings -----
	// ****************************

	if err = File.Section("session").MapTo(&Session); err != nil {
		return errors.Wrap(err, "mapping [session] section")
	}

	// ***************************
	// ----- SMTP settings -----
	// ***************************
	if err = File.Section("smtp").MapTo(&Smtp); err != nil {
		return errors.Wrap(err, "mapping [smtp] section")
	}

	// ***************************
	// ----- Pop3 settings -----
	// ***************************
	if err = File.Section("pop3").MapTo(&Pop3); err != nil {
		return errors.Wrap(err, "mapping [pop] section")
	}

	// ***************************
	// ----- Imap settings -----
	// ***************************
	if err = File.Section("imap").MapTo(&Imap); err != nil {
		return errors.Wrap(err, "mapping [imap] section")
	}

	// ***************************
	// ----- Rspamd settings -----
	// ***************************
	if err = File.Section("rspamd").MapTo(&Rspamd); err != nil {
		return errors.Wrap(err, "mapping [rspamd] section")
	}

	// *****************************
	// ----- Security settings -----
	// *****************************

	if err = File.Section("security").MapTo(&Security); err != nil {
		return errors.Wrap(err, "mapping [security] section")
	}

	// ***************************
	// ----- i18n settings -----
	// ***************************
	I18n = new(i18nConf)
	if err = File.Section("i18n").MapTo(&I18n); err != nil {
		return errors.Wrap(err, "mapping [i18n] section")
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
