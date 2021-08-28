package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func IndexWeb(c *gin.Context) {
	c.String(http.StatusOK, "hello world")
}

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/", IndexWeb)

	v1 := r.Group("v1")
	{
		v1.POST("/login", UserLogin)
	}

	return r
}

func Start(port int) {
	r := SetupRouter()

	//监听端口默认为8080
	listen_port := fmt.Sprintf(":%d", port)
	r.Run(listen_port)
}
