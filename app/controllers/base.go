package controllers

import (
	"github.com/astaxie/beego"
	// "strconv"
	// "strings"
	// "time"
)

const (
	MSG_OK  = 0
	MSG_ERR = -1
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

func (t *BaseController) makeJwt(userid string, username string) string {
	return ""
}

func (t *BaseController) wailJwt(userid string, username string) string {
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
