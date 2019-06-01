package controllers

import (
	"fmt"
	// "github.com/astaxie/beego"
	// "github.com/dgrijalva/jwt-go/request"
	"github.com/midoks/imail/app/models"
)

//UserController ...
type UserController struct {
	BaseController
}

//Login ...
func (t *UserController) In() {

	username := t.GetString("username")
	password := t.GetString("password")
	fmt.Println(username, password)

	tokenString := t.makeJwt("d", username)

	t.retOk(tokenString)
}

//Login ...
func (t *UserController) Out() {
	t.retOk("ok")
}

func (t *UserController) Info() {
	id, _ := t.GetInt("id")
	// fmt.Println(id)
	row, _ := models.UserGetById(id)
	fmt.Println(row)
	t.retOk(row)
}
