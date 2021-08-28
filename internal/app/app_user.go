package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	// "github.com/midoks/imail/internal/db"
)

func UserRegister(c *gin.Context) {

}

func GetCodeByName(c *gin.Context) {
	name := c.PostForm("name")

	fmt.Println(name)
}

func UserLogin(c *gin.Context) {
	name := c.Query("name")
	password := c.Query("password")

	fmt.Println("UserLogin:", name, password)

	// b, _ := db.LoginByUserPassword(name, password)
	// fmt.Println(b, id)
	if false {
		c.JSON(200, gin.H{"msg": "login success!"})
	} else {
		c.JSON(200, gin.H{"msg": "login error!"})
	}
}
