package app

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func IndexWeb(c *gin.Context) {
	c.String(http.StatusOK, "hello word")
}

func LoginWeb(c *gin.Context) {

}
