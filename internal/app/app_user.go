package app

import (
	// "fmt"
	// "github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/libs"
)

func UserRegister(c *gin.Context) {

}

func GetUserCode(c *gin.Context) {
	rand := libs.RandString(10)
	token := libs.Md5str(rand)

	name := c.Query("name")
	if name == "" {
		c.JSON(200, gin.H{"code": -1, "rand": rand, "token": token})
		return
	}

	db.UserLoginVerifyAdd(name, rand, token)
	c.JSON(200, gin.H{"code": "0", "rand": rand, "token": token})
}

func UserLogin(c *gin.Context) {
	name := c.PostForm("name")
	password := c.PostForm("password")
	token := c.PostForm("token")

	// fmt.Println("UserLogin:", name, password, token)

	r, _ := db.UserLoginVerifyGet(name)

	sessRand := r.Rand
	sessToken := r.Token

	if sessRand == "" {
		c.JSON(200, gin.H{"code": "-1", "msg": "need to get code!"})
		return
	}

	if sessToken != token {
		c.JSON(200, gin.H{"code": "-1", "msg": "token is error!"})
		return
	}

	b, _ := db.LoginByUserPassword(name, password, sessRand)
	loginToken := libs.Md5str(libs.RandString(10))

	db.UserUpdateTokenGetByName(name, loginToken)
	if b {
		c.JSON(200, gin.H{"code": "0", "msg": "login success!", "token": loginToken})
	} else {
		c.JSON(200, gin.H{"code": "-1", "msg": "login fail!"})
	}
}
