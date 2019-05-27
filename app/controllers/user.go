package controllers

import (
	"fmt"
	"github.com/midoks/imail/app/models"
)

// "github.com/astaxie/beego"
// "strconv"
// "strings"
// "time"

//UserController ...
type UserController struct {
	BaseController
}

//Login ...
func (t *UserController) In() {

	t.retOk("ok")
}

//Login ...
func (t *UserController) Out() {
	t.retOk("ok")
}

func (t *UserController) Info() {
	id, _ := t.GetInt("id")
	fmt.Println(id)
	row, _ := models.UserGetById(id)
	fmt.Println(row)
	t.retOk(row)
}
