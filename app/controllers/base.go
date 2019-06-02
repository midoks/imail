package controllers

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/dgrijalva/jwt-go"
	"time"
)

const (
	MSG_OK  = 0
	MSG_ERR = -1
)

const (
	SecretKey = "imail"
)

/**
 * BaseController ...
 * base struct ...
 */
type BaseController struct {
	beego.Controller
	controllerName string
	actionName     string
	pageSize       int

	// xsrf data
	_xsrfToken string
	XSRFExpire int
	EnableXSRF bool
}

// 输出json ...
func (t *BaseController) retJson(out interface{}) {
	t.Data["json"] = out
	t.ServeJSON()
	t.StopRun()
}

func (t *BaseController) makeJwt(userid int64, username string) string {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := make(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(1)).Unix()
	claims["iat"] = time.Now().Unix()
	claims["id"] = userid
	claims["username"] = username
	token.Claims = claims
	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		fmt.Println("Error extracting the key")
		return ""
	}
	return tokenString
}

func (t *BaseController) wailJwt(token string) string {
	// fmt.Println(tokenString)

	// claims2, err := jwt.Parse(tokenString, func(*jwt.Token) (interface{}, error) {
	// 	return []byte(SecretKey), nil
	// })

	// if err != nil {
	// 	fmt.Println("转换为jwt claims失败.", err)
	// }

	// fmt.Println(claims2)
	return ""
}

func (t *BaseController) retResult(code int, msg interface{}, data ...interface{}) {
	out := make(map[string]interface{})
	out["code"] = code
	out["msg"] = msg

	if len(data) > 0 {
		out["data"] = data
	}

	t.retJson(out)
}

func (t *BaseController) retOk(msg interface{}, data ...interface{}) {
	t.retResult(MSG_OK, msg, data...)
}

func (t *BaseController) retFail(msg interface{}, data ...interface{}) {
	t.retResult(MSG_ERR, msg, data...)
}

// 是否POST提交 ...
func (t *BaseController) isPost() bool {
	return t.Ctx.Request.Method == "POST"
}
