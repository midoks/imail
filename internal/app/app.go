package app

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/midoks/imail/internal/config"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/denyip"
	"github.com/midoks/imail/internal/log"
	uuid "github.com/satori/go.uuid"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

var checker *denyip.Checker

func FixTestMiddleware() {
	if !db.CheckDb() {
		os.MkdirAll("data", 0777)
		err := config.Load("../../conf/app.defined.conf")
		if err != nil {
			panic("config file load err")
		}

		log.Init()
		db.Init()
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

const xForwardedFor = "X-Forwarded-For"

func getRemoteIP(req *http.Request) []string {
	var ipList []string

	xff := req.Header.Get(xForwardedFor)
	xffs := strings.Split(xff, ",")

	for i := len(xffs) - 1; i >= 0; i-- {
		xffsTrim := strings.TrimSpace(xffs[i])

		if len(xffsTrim) > 0 {
			ipList = append(ipList, xffsTrim)
		}
	}

	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		remoteAddrTrim := strings.TrimSpace(req.RemoteAddr)
		if len(remoteAddrTrim) > 0 {
			ipList = append(ipList, remoteAddrTrim)
		}
	} else {
		ipTrim := strings.TrimSpace(host)
		if len(ipTrim) > 0 {
			ipList = append(ipList, ipTrim)
		}
	}

	return ipList
}

func IPWhiteMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ipWhiteList := strings.Split(config.GetString("http.ip_white", "*"), ",")
		if !config.InSliceString("*", ipWhiteList) && len(ipWhiteList) != 0 {
			reqIPAddr := getRemoteIP(c.Request)
			reeIPadLenOffset := len(reqIPAddr) - 1
			for i := reeIPadLenOffset; i >= 0; i-- {
				err := checker.IsAuthorized(reqIPAddr[i])
				if err != nil {
					log.Error(err)
					c.String(http.StatusForbidden, err.Error())
					return
				}
			}
		}
		c.Next()
	}
}

func IndexWeb(c *gin.Context) {
	c.String(http.StatusOK, "hello world")
}

func SetupRouter() *gin.Engine {
	r := gin.Default()

	FixTestMiddleware()
	r.Use(RequestIDMiddleware(), LogMiddleware(), IPWhiteMiddleware())

	if b, _ := config.GetBool("redis.enable", false); b {
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
	}

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
	ipWhiteList := strings.Split(config.GetString("http.ip_white", "*"), ",")
	if !config.InSliceString("*", ipWhiteList) && len(ipWhiteList) != 0 {
		var err error
		checker, err = denyip.NewChecker(ipWhiteList)
		if err != nil {
			log.Fatal(err)
		}
	}
	r := SetupRouter()

	//Listening port
	listen_port := fmt.Sprintf(":%d", port)
	r.Run(listen_port)
}
