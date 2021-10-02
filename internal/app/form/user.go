package form

import (
// "mime/multipart"

// "github.com/go-macaron/binding"
// "gopkg.in/macaron.v1"
)

type Register struct {
	UserName string `binding:"Required;AlphaDashDot;MaxSize(35)"`
	Password string `binding:"Required;MaxSize(255)"`
	Retype   string
}
