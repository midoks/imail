package routers

import (
	"fmt"

	"github.com/astaxie/beego"
	"github.com/midoks/imail/app/controllers"
	"github.com/midoks/imail/app/models"

	// "html/template"
	"net/http"
)

func page_not_found(rw http.ResponseWriter, r *http.Request) {
	// t, _ := template.New("404.html").ParseFiles(beego.BConfig.WebConfig.ViewsPath + "/404.html")
	data := make(map[string]interface{})
	data["content"] = "page not found"
	// this.Redirect(url, 302)
	rw.Write([]byte("page not found"))
	// t.Execute(rw, data)
}

func init() {
	fmt.Println("routers init")
	models.Init()

	//错误页面设置
	beego.ErrorHandler("404", page_not_found)

	beego.AutoRouter(&controllers.UserController{})

	//v1
	ns := beego.NewNamespace("/v1", beego.NSAutoRouter(&controllers.UserController{}))
	beego.AddNamespace(ns)
}
