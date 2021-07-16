package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func Start(port int) {

	r := gin.Default()
	r.GET("/", IndexWeb)
	r.GET("/login", LoginWeb)

	//监听端口默认为8080
	listen_port := fmt.Sprintf(":%d", port)
	// fmt.Println(listen_port)
	r.Run(listen_port)
}
