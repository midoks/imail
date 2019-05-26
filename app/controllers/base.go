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

// 输出json
func (this *BaseController) retJson(out interface{}) {
	this.Data["json"] = out
	this.ServeJSON()
	this.StopRun()
}

func (this *BaseController) retResult(code int, msg interface{}, data ...interface{}) {
	out := make(map[string]interface{})
	out["code"] = code
	out["msg"] = msg

	if len(data) > 0 {
		out["data"] = data
	}

	this.retJson(out)
}

func (this *BaseController) retOk(msg interface{}, data ...interface{}) {
	this.retResult(MSG_OK, msg, data...)
}

func (this *BaseController) retFail(msg interface{}, data ...interface{}) {
	this.retResult(MSG_ERR, msg, data...)
}
