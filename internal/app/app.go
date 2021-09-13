package app

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/midoks/imail/internal/config"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/log"
	"net/http"
	"time"
)

func FixTestMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !db.CheckDb() {
			err := config.Load("../../conf/app.conf")
			if err != nil {
				panic("config file load err")
			}

			log.Init()
			db.Init()
		}
	}
}

func LogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		//Processing requests
		c.Next()

		endTime := time.Now()
		latencyTime := endTime.Sub(startTime)
		reqMethod := c.Request.Method
		reqUrl := c.Request.RequestURI
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		logger := log.GetLogger()
		logger.Infof("| %3d | %13v | %15s | %s | %s |",
			statusCode,
			latencyTime,
			clientIP,
			reqMethod,
			reqUrl,
		)
	}
}

func IndexWeb(c *gin.Context) {
	c.String(http.StatusOK, "hello world")
}

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(FixTestMiddleware())
	r.Use(LogMiddleware())

	store, err := redis.NewStore(10, "tcp", "127.0.0.1:6379", "", []byte("secret"))
	if err != nil {
		store = cookie.NewStore([]byte("SESSION_SECRET"))
	}
	store.Options(sessions.Options{MaxAge: 60 * 60})
	r.Use(sessions.Sessions("sessionid", store))

	r.GET("/", IndexWeb)
	v1 := r.Group("v1")
	{
		v1.GET("/get_code", GetUserCode)
		v1.POST("/update_user_code", UpdateUserCodeByName)
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
