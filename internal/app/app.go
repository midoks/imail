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
	uuid "github.com/satori/go.uuid"
	"net/http"
	"time"
)

func FixTestMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !db.CheckDb() {
			err := config.Load("conf/app.defined.conf")
			if err != nil {
				panic("config file load err")
			}

			log.Init()
			db.Init()
		}
	}
}

// LogMiddleware 访问日志中间件
func LogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 捕抓异常
		defer func() {
			if err := recover(); err != nil {
				log.Error(err)
			}
		}()
		startTime := time.Now()

		//Processing requests
		c.Next()
		// 结束时间
		endTime := time.Now()
		// 执行时间
		latencyTime := endTime.Sub(startTime)
		// 请求方式
		reqMethod := c.Request.Method
		// 请求路由
		reqUrl := c.Request.RequestURI
		// 请求ID
		requestID := c.Request.Header.Get("X-Request-Id")
		// 状态码
		statusCode := c.Writer.Status()
		// 请求IP
		clientIP := c.ClientIP()
		// 请求协议
		proto := c.Request.Proto

		logger := log.GetLogger()
		logger.Infof("| %3d | %13v | %15s | %s | %s | %s | %s |",
			statusCode,
			latencyTime,
			clientIP,
			proto,
			reqMethod,
			requestID,
			reqUrl,
		)
	}
}

const xRequestIDKey = "X-Request-ID"

// RequestIDMiddleware 请求ID 中间件
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		u4 := uuid.NewV4()
		xRequestID := u4.String()
		c.Request.Header.Set(xRequestIDKey, xRequestID)
		c.Writer.Header().Set(xRequestIDKey, xRequestID)
		c.Set(xRequestIDKey, xRequestID)
		c.Next()
	}
}

func IndexWeb(c *gin.Context) {
	c.String(http.StatusOK, "hello world")
}

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(FixTestMiddleware(), RequestIDMiddleware(), LogMiddleware())

	store, err := redis.NewStoreWithDB(
		10, "tcp",
		config.GetString("redis.address", "127.0.0.1:6379"),
		config.GetString("redis.password", ""),
		config.GetString("redis.db", "0"),
		[]byte("secret"),
	)
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

	//Listening port
	listen_port := fmt.Sprintf(":%d", port)
	r.Run(listen_port)
}
