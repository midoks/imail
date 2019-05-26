package controllers

import (
// "github.com/astaxie/beego"
// "strconv"
// "strings"
// "time"
)

type UserController struct {
	BaseController
}

func (this *UserController) login() {
	this.retOk("ok")
}
