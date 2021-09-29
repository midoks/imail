package app

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/go-macaron/gzip"
	"github.com/midoks/imail/internal/app/template"
	"github.com/midoks/imail/internal/config"
	"github.com/midoks/imail/internal/denyip"
	"github.com/midoks/imail/internal/log"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/macaron.v1"
	"net"
	"net/http"
	"strings"
	"time"
)

var checker *denyip.Checker

// LogMiddleware
func LogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// catch
		defer func() {
			if err := recover(); err != nil {
				log.Error(err)
			}
		}()

		startTime := time.Now()
		// processing requests
		c.Next()
		endTime := time.Now()

		// run time
		latencyTime := endTime.Sub(startTime)
		// method
		reqMethod := c.Request.Method
		// uri
		reqUrl := c.Request.RequestURI
		// X-Request-Id
		requestID := c.Request.Header.Get("X-Request-Id")
		// status code
		statusCode := c.Writer.Status()
		// request ip
		clientIP := c.ClientIP()
		// request protocol
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
	c.HTML(http.StatusOK, "index.tmpl", gin.H{})
}

func SetupRouter() *gin.Engine {
	r := gin.Default()

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

	// r.LoadHTMLGlob("templates/*")
	r.LoadHTMLGlob("templates/**/*")
	r.GET("/", IndexWeb)
	v1 := r.Group("v1")
	{
		v1.GET("/get_code", GetUserCode)
		v1.POST("/update_user_code", UpdateUserCodeByName)
		v1.POST("/login", UserLogin)
	}

	return r
}

func Start2(port int) {
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

func newMacaron() *macaron.Macaron {
	m := macaron.New()
	m.Use(macaron.Logger())
	m.Use(gzip.Gziper())
	m.Use(macaron.Logger())
	m.Use(macaron.Recovery())

	m.Use(macaron.Static("public"))

	opt := macaron.Renderer(macaron.RenderOptions{
		Directory: "templates/default",
		Funcs:     template.FuncMap(),
	})
	m.Use(opt)

	return m
}

func setRouter(m *macaron.Macaron) *macaron.Macaron {

	m.Group("", func() {
		m.Get("/", func(ctx *macaron.Context) {
			ctx.HTML(200, "home")
		})

		m.Get("/install", func(ctx *macaron.Context) {
			ctx.HTML(200, "install")
		})
	})
	return m
}

func Start(port int) {
	m := newMacaron()
	m = setRouter(m)
	m.Run(port)
}
