// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package template

import (
	"fmt"
	"html/template"
	"mime"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/editorconfig/editorconfig-core-go/v2"
	"github.com/microcosm-cc/bluemonday"
	"github.com/midoks/imail/internal/conf"
	"github.com/unknwon/i18n"
)

var (
	funcMap     []template.FuncMap
	funcMapOnce sync.Once
)

// FuncMap returns a list of user-defined template functions.
func FuncMap() []template.FuncMap {
	funcMapOnce.Do(func() {
		funcMap = []template.FuncMap{map[string]interface{}{
			"BuildCommit": func() string {

				t := time.Now().Unix()
				s := strconv.FormatInt(t, 10)
				return s
				// return conf.BuildCommit
			},

			"Year": func() int {
				return time.Now().Year()
			},
			"AppSubURL": func() string {
				return conf.Web.Subpath
			},
			"AppName": func() string {
				return conf.App.Name
			},
			"AppVer": func() string {
				return conf.App.Version
			},
			"AppDomain": func() string {
				return conf.Web.Domain
			},

			"Safe":        Safe,
			"Str2HTML":    Str2HTML,
			"Sanitize":    bluemonday.UGCPolicy().Sanitize,
			"NewLine2br":  NewLine2br,
			"EscapePound": EscapePound,
			"Add": func(a, b int) int {
				return a + b
			},

			"SubStr": func(str string, start, length int) string {
				if len(str) == 0 {
					return ""
				}
				end := start + length
				if length == -1 {
					end = len(str)
				}
				if len(str) < end {
					return str
				}
				return str[start:end]
			},
			"ShowFooterTemplateLoadTime": func() bool {
				return conf.Other.ShowFooterTemplateLoadTime
			},
			"LoadTimes": func(startTime time.Time) string {
				return fmt.Sprint(time.Since(startTime).Nanoseconds()/1e6) + "ms"
			},
			"Join": strings.Join,
			"DateFmtLong": func(t time.Time) string {
				return t.Format(time.RFC1123Z)
			},
			"DateFmtShort": func(t time.Time) string {
				fmt.Println(t)
				return t.Format("Jan 02, 2006")
			},

			"DateFmtMail":      DateFmtMail,
			"DateInt64FmtMail": DateInt64FmtMail,

			"FilenameIsImage": func(filename string) bool {
				mimeType := mime.TypeByExtension(filepath.Ext(filename))
				return strings.HasPrefix(mimeType, "image/")
			},
			"TabSizeClass": func(ec *editorconfig.Editorconfig, filename string) string {
				if ec != nil {
					def, err := ec.GetDefinitionForFilename(filename)
					if err == nil && def.TabWidth > 0 {
						return fmt.Sprintf("tab-size-%d", def.TabWidth)
					}
				}
				return "tab-size-8"
			},
		}}
	})
	return funcMap
}

func Safe(raw string) template.HTML {
	return template.HTML(raw)
}

func Str2HTML(raw string) template.HTML {
	return template.HTML(bluemonday.UGCPolicy().Sanitize(raw))
}

// NewLine2br simply replaces "\n" to "<br>".
func NewLine2br(raw string) string {
	return strings.Replace(raw, "\n", "<br>", -1)
}

// TODO: Use url.Escape.
func EscapePound(str string) string {
	return strings.NewReplacer("%", "%25", "#", "%23", " ", "%20", "?", "%3F").Replace(str)
}

func DateFmtMail(t time.Time, lang string) string {
	n := time.Now()

	in := t.Format("2006-01-02")
	now := n.Format("2006-01-02")

	if in == now {
		return t.Format("15:04")
	}
	in2, _ := time.Parse("2006-01-02 15:04:05", in+" 00:00:00")
	now2, _ := time.Parse("2006-01-02 15:04:05", now+" 00:00:00")
	if in2.Unix()+86400 == now2.Unix() {
		return i18n.Tr(lang, "common.yesterday")
	} else {
		return t.Format("2006-01-02")
	}
}

func DateInt64FmtMail(t int64, lang string) string {
	n := time.Now()

	in := time.Unix(t, 0).Format("2006-01-02")
	now := n.Format("2006-01-02")

	if in == now {
		return time.Unix(t, 0).Format("15:04")
	}
	in2, _ := time.Parse("2006-01-02 15:04:05", in+" 00:00:00")
	now2, _ := time.Parse("2006-01-02 15:04:05", now+" 00:00:00")
	if in2.Unix()+86400 == now2.Unix() {
		return i18n.Tr(lang, "common.yesterday")
	} else {
		return time.Unix(t, 0).Format("2006-01-02")
	}
}
