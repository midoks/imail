package app

import (
	"github.com/astaxie/beego"
	_ "github.com/midoks/imail/app/routers"
)

func Start() {
	beego.Run()
}
