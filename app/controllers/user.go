package controllers

import (
	"fmt"
	// "github.com/astaxie/beego"
	"github.com/dgrijalva/jwt-go"
	// "github.com/dgrijalva/jwt-go/request"
	"github.com/midoks/imail/app/models"
	"time"
)

// "github.com/astaxie/beego"
// "strconv"
// "strings"

//UserController ...
type UserController struct {
	BaseController
}

//Login ...
func (t *UserController) In() {

	username := t.GetString("username")
	fmt.Println(username)

	token := jwt.New(jwt.SigningMethodHS256)
	claims := make(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(1)).Unix()
	claims["iat"] = time.Now().Unix()
	claims["id"] = "1"
	claims["username"] = "midoks"

	token.Claims = claims

	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {

		fmt.Println("Error extracting the key")
		// fatal(err)
	}

	// fmt.Println(tokenString)

	// claims2, err := jwt.Parse(tokenString, func(*jwt.Token) (interface{}, error) {
	// 	return []byte(SecretKey), nil
	// })

	// if err != nil {
	// 	fmt.Println("转换为jwt claims失败.", err)
	// }

	// fmt.Println(claims2)

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
