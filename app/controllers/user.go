package controllers

import (
	"fmt"
	"github.com/midoks/imail/app/models"
	"github.com/midoks/imail/libs"
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

	info, err := models.UserGetByName(username)
	if err != nil {
		errInfo := fmt.Sprintf("username %s does not exist!", username)
		t.retFail(errInfo)
	}

	loginPasswd := libs.Md5str(password)
	if loginPasswd != info.Password {
		t.retFail("landing failed!")
	}

	tokenString := t.makeJwt(info.Id, info.Name)

	data := make(map[string]interface{})
	data["token"] = tokenString
	t.retOk("landing success!", data)
}

func (t *UserController) Info() {
	id, _ := t.GetInt("id")
	// fmt.Println(id)
	row, _ := models.UserGetById(id)
	fmt.Println(row)
	t.retOk(row)
}
