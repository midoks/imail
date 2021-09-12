package log

import (
	"fmt"
	// "github.com/gin-gonic/gin"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/midoks/imail/internal/config"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"strings"
	"time"
)

var (
	logFilePath = "./logs"
	logFileName = "system.log"
	logger      *logrus.Logger
)

func Init() {
	fileName := path.Join(logFilePath, logFileName)

	src, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		fmt.Println("err", err)
	}
	logger = logrus.New()

	logger.SetFormatter(&logrus.TextFormatter{})
	logger.Out = src

	if config.IsLoaded() {
		runmode := config.GetString("runmode", "dev")
		if strings.EqualFold(runmode, "dev") {
			logger.SetLevel(logrus.DebugLevel)
		} else {
			logger.SetLevel(logrus.InfoLevel)
		}
	} else {
		logger.SetLevel(logrus.DebugLevel)
	}

	// 设置 rotatelogs
	logWriter, err := rotatelogs.New(
		// 分割后的文件名称
		fileName+".%Y%m%d.log",
		// 生成软链，指向最新日志文件
		rotatelogs.WithLinkName(fileName),
		// 设置最大保存时间(7天)
		rotatelogs.WithMaxAge(7*24*time.Hour),
		// 设置日志切割时间间隔(1天)
		rotatelogs.WithRotationTime(1*time.Minute),
	)

	writeMap := lfshook.WriterMap{
		logrus.TraceLevel: logWriter,
		logrus.InfoLevel:  logWriter,
		logrus.FatalLevel: logWriter,
		logrus.DebugLevel: logWriter,
		logrus.WarnLevel:  logWriter,
		logrus.ErrorLevel: logWriter,
		logrus.PanicLevel: logWriter,
	}

	logger.AddHook(lfshook.NewHook(writeMap, &logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	}))

	// return func(c *gin.Context) {
	// 	//开始时间
	// 	startTime := time.Now()
	// 	//处理请求
	// 	c.Next()
	// 	//结束时间
	// 	endTime := time.Now()
	// 	// 执行时间
	// 	latencyTime := endTime.Sub(startTime)
	// 	//请求方式
	// 	reqMethod := c.Request.Method
	// 	//请求路由
	// 	reqUrl := c.Request.RequestURI
	// 	//状态码
	// 	statusCode := c.Writer.Status()
	// 	//请求ip
	// 	clientIP := c.ClientIP()

	// 	// 日志格式
	// 	logger.WithFields(logrus.Fields{
	// 		"status_code":  statusCode,
	// 		"latency_time": latencyTime,
	// 		"client_ip":    clientIP,
	// 		"req_method":   reqMethod,
	// 		"req_uri":      reqUrl,
	// 	}).Info()
	// }

	// log debug
	// logger.WithFields().Info()
	// logger.WithFields(logrus.Fields{
	// 	"animal": "walrus",
	// }).Info("A walrus appears")

}

func Trace(args ...interface{}) {
	logger.Trace(args...)
}

func Debug(args ...interface{}) {
	logger.Debug(args...)
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Warn(args ...interface{}) {
	logger.Warn(args...)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}

func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

func Panic(args ...interface{}) {
	logger.Panic(args...)
}

func Tracef(format string, args ...interface{}) {
	logger.Tracef(format, args...)
}

func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	logger.Panicf(format, args...)
}
