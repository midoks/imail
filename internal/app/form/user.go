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

type ChangePassword struct {
	OldPassword string `binding:"Required;MinSize(1);MaxSize(255)"`
	Password    string `binding:"Required;MaxSize(255)"`
	Retype      string
}

type UpdateProfile struct {
	Nick string `binding:"Required;AlphaDashDot;MaxSize(35)"`
}
