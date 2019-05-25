package routers

import (
	"github.com/astaxie/beego"
	// "github.com/astaxie/beego/context"
	"github.com/midoks/imail/web/controllers"
)

func init() {
	// beego.Get("/", func(ctx *context.Context) {
	// 	ctx.Output.Body([]byte("hello world"))
	// })

	beego.Router("/", &controllers.Index{}, "*:Index")
}
