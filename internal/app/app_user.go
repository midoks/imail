package app

import (
	// "fmt"
	// "github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/tools"
)

func UserRegister(c *gin.Context) {

}

func GetUserCode(c *gin.Context) {
	rand := tools.RandString(10)
	token := tools.Md5(rand)

	name := c.Query("name")
	if name == "" {
		c.JSON(200, gin.H{"code": -1, "rand": rand, "token": token})
		return
	}

	db.UserLoginVerifyAdd(name, rand, token)
	c.JSON(200, gin.H{"code": "0", "rand": rand, "token": token})
}

// update user code for mail client
func UpdateUserCodeByName(c *gin.Context) {
	name := c.PostForm("name")

	if name == "" {
		c.JSON(200, gin.H{"code": "-1", "msg": "name cannot be empty！"})
	}

	if db.UserCheckIsExist(name) {
		rand := tools.RandString(10)
		db.UserUpdateCodeGetByName(name, rand)
		c.JSON(200, gin.H{"code": "0", "co": rand})
	} else {
		c.JSON(200, gin.H{"code": "-2", "msg": "name does not exist！"})
	}

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

	b, _ := db.LoginByUserPassword(name, password)
	loginToken := tools.Md5(tools.RandString(10))

	db.UserUpdateTokenGetByName(name, loginToken)
	if b {
		c.JSON(200, gin.H{"code": "0", "msg": "login success!", "token": loginToken})
	} else {
		c.JSON(200, gin.H{"code": "-1", "msg": "login fail!"})
	}
}
