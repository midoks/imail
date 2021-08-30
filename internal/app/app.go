package app

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	// "github.com/gin-contrib/sessions/cookie"
	// "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/login") ||
			strings.HasPrefix(c.Request.URL.Path, "/signup") {
			return
		}
		if strings.HasPrefix(c.Request.URL.Path, "/static") {
			return
		}

		session := sessions.Default(c)
		bunny := session.Get("authenticated")
		if bunny == nil || bunny == false {
			c.Redirect(http.StatusPermanentRedirect, "/")
		} else {
			c.Next()
		}

	}
}

func IndexWeb(c *gin.Context) {
	c.String(http.StatusOK, "hello world")
}

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// store := cookie.NewStore([]byte("SESSION_SECRET"))

	// store, _ := redis.NewStore(10, "tcp", "localhost:6379", "", []byte("secret"))
	// store.Options(sessions.Options{MaxAge: 60 * 10})
	// r.Use(sessions.Sessions("sessionid", store))

	r.GET("/", IndexWeb)

	v1 := r.Group("v1")
	{
		v1.GET("/get_code", GetUserCode)
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
