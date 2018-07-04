package controllers

import (
	"fmt"
	"github.com/astaxie/beego"
)

type CommonController struct {
	beego.Controller
}

func (this *CommonController) Prepare() {
	fmt.Println("123123")
}

func (this *CommonController) Get() {
	fmt.Println("123123")
}
