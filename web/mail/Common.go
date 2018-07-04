package web

import (
	"github.com/astaxie/beego"
)

type CommonController struct {
	beego.Controller
}

func (this *CommonController) Prepare() {
	fmt.Println("123123")
}
