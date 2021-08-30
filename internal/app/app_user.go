package app

import (
	"fmt"
	"github.com/gin-contrib/sessions"
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

	fmt.Println("name:", name)

	if name == "" {
		c.JSON(200, gin.H{"code": -1, "rand": rand, "token": token})
		return
	}

	db.UserLoginVerifyAdd(name, rand, token)

	// sess := sessions.Default(c)
	// sess.Set("rand", rand)
	// sess.Set("token", token)
	// sess.Save()

	c.JSON(200, gin.H{"code": "0", "rand": rand, "token": token})

}

func UserLogin(c *gin.Context) {
	name := c.PostForm("name")
	password := c.PostForm("password")
	token := c.PostForm("token")

	// fmt.Println("UserLogin:", name, password, token)

	sess := sessions.Default(c)
	sessRand := sess.Get("rand")
	sessToken := sess.Get("token")

	if sessRand == nil {
		c.JSON(200, gin.H{"code": "-1", "msg": "need to get code!"})
		return
	}

	if sessToken.(string) != token {
		c.JSON(200, gin.H{"code": "-1", "msg": "token is error!"})
		return
	}

	b, _ := db.LoginByUserPassword(name, password, sessRand.(string))
	if b {
		c.JSON(200, gin.H{"code": "0", "msg": "login success!"})
	} else {
		c.JSON(200, gin.H{"code": "-1", "msg": "login fail!"})
	}
}
