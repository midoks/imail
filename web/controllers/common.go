package controllers

import (
	"fmt"
	"github.com/astaxie/beego"
	"strconv"
	"strings"
	"time"
)

// const (
// 	MSG_OK  = 0
// 	MSG_ERR = -1
// )

type Common struct {
	beego.Controller
	controllerName string
	actionName     string
	pageSize       int

	// xsrf data
	_xsrfToken string
	XSRFExpire int
	EnableXSRF bool
}

// func (this *CommonController) uLog(behavior string) {
// 	models.LogAdd(this.user.Id, 1, behavior)
// }

func (this *Index) Prepare() {
	this.initData()
}

func (this *Common) dLog(behavior string) {
	// models.DebugAdd(1, behavior)
}

func (this *Common) D(args ...string) {
	if beego.AppConfig.String("runmode") == "dev" {
		for i := 0; i < len(args); i++ {
			this.Ctx.WriteString(args[i])
		}
	}
}

func (this *Common) P(args ...string) {
	if beego.AppConfig.String("runmode") == "dev" {
		for i := 0; i < len(args); i++ {
			fmt.Println(args[i])
		}
	}
}

func (this *Common) initXSRF() {
	this.EnableXSRF = true
	this._xsrfToken = "61oETzKXQAGaYdkL5gEmGeJJFuYh7EQnp2XdTP1o"
	this.XSRFExpire = 3600 //过期时间，默认1小时
}

func (this *Common) initData() {

	this.Data["pageStartTime"] = time.Now()
	this.pageSize = 20
	controllerName, actionName := this.GetControllerAndAction()
	this.controllerName = strings.ToLower(controllerName)
	this.actionName = strings.ToLower(actionName)

	this.Data["version"] = beego.AppConfig.String("version")
	this.Data["siteName"] = beego.AppConfig.String("site.name")
	this.Data["curRoute"] = this.controllerName + "." + this.actionName
	this.Data["curController"] = this.controllerName
	this.Data["curAction"] = this.actionName
}

func (this *Common) isIntInList(check int, list string) (out bool) {
	out = false
	numList := strings.Split(list, ",")
	for i := 0; i < len(numList); i++ {
		if numList[i] == strconv.Itoa(check) {
			out = true
		}
	}
	return out
}

// 是否POST提交
func (this *Common) isPost() bool {
	return this.Ctx.Request.Method == "POST"
}

// 重定向
func (this *Common) redirect(url string) {
	this.Redirect(url, 302)
	this.StopRun()
}

//获取用户IP地址
func (this *Common) getClientIp() string {
	s := strings.Split(this.Ctx.Request.RemoteAddr, ":")
	return s[0]
}

//渲染模版
func (this *Common) display(tpl ...string) {
	var tplname string
	if len(tpl) == 1 {
		tplname = tpl[0] + ".html"
	} else if len(tpl) == 2 {
		tplname = tpl[0] + "/" + tpl[1] + ".html"
	} else {
		tplname = this.controllerName + "/" + this.actionName + ".html"
	}

	this.Layout = "layout/index.html"
	this.TplName = tplname
}

// 输出json
func (this *Common) retJson(out interface{}) {
	this.Data["json"] = out
	this.ServeJSON()
	this.StopRun()
}

func (this *Common) retResult(code int, msg interface{}, data ...interface{}) {
	out := make(map[string]interface{})
	out["code"] = code
	out["msg"] = msg

	if len(data) > 0 {
		out["data"] = data
	}

	this.retJson(out)
}

func (this *Common) retOk(msg interface{}, data ...interface{}) {
	this.retResult(MSG_OK, msg, data...)
}

func (this *Common) retFail(msg interface{}, data ...interface{}) {
	this.retResult(MSG_ERR, msg, data...)
}
