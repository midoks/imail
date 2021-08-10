package app

import (
	// "fmt"
	"github.com/gin-gonic/gin"
	"github.com/midoks/imail/internal/db"
)

func UserRegister(c *gin.Context) {

}

func UserLogin(c *gin.Context) {
	name := c.PostForm("name")
	password := c.PostForm("password")

	b, _ := db.LoginByUserPassword(name, password)
	// fmt.Println(b, id)
	if b {
		c.JSON(200, gin.H{"msg": "login success!"})
	} else {
		c.JSON(200, gin.H{"msg": "login error!"})
	}

}
