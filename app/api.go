package app

import (
	"github.com/astaxie/beego"
	_ "github.com/midoks/imail/app/routers"
)

func Start(port int) {

	beego.BConfig.Listen.HTTPPort = port
	beego.Run()
}
