// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package template

import (
	"fmt"
	"html/template"
	"mime"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/editorconfig/editorconfig-core-go/v2"
	"github.com/microcosm-cc/bluemonday"
	"github.com/midoks/imail/internal/conf"
)

var (
	funcMap     []template.FuncMap
	funcMapOnce sync.Once
)

// FuncMap returns a list of user-defined template functions.
func FuncMap() []template.FuncMap {
	funcMapOnce.Do(func() {
		funcMap = []template.FuncMap{map[string]interface{}{
			"Year": func() int {
				return time.Now().Year()
			},
			"AppSubURL": func() string {
				return conf.Web.Subpath
			},
			"AppName": func() string {
				return conf.App.Name
			},
			"LoadTimes": func(startTime time.Time) string {
				return fmt.Sprint(time.Since(startTime).Nanoseconds()/1e6) + "ms"
			},
			"Safe":       Safe,
			"Str2HTML":   Str2HTML,
			"Sanitize":   bluemonday.UGCPolicy().Sanitize,
			"NewLine2br": NewLine2br,
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
			"Join": strings.Join,

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
