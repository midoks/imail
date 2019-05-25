package api

import "github.com/gin-gonic/gin"

func Start() {
	router := gin.Default()

	v1 := router.Group("/v1")
	{
		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})
	}

	router.Run(":8091") // listen and serve on 0.0.0.0:8080
}
