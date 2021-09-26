package log

import (
	"fmt"
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
		fmt.Println("log error", err)
	}
	logger = logrus.New()

	if config.IsLoaded() {
		format := config.GetString("log.format", "json")
		if strings.EqualFold(format, "json") {
			logger.SetFormatter(&logrus.JSONFormatter{})
		} else if strings.EqualFold(format, "text") {
			logger.SetFormatter(&logrus.TextFormatter{})
		} else {
			logger.SetFormatter(&logrus.TextFormatter{})
		}

		runmode := config.GetString("runmode", "dev")
		if strings.EqualFold(runmode, "dev") {
			logger.SetLevel(logrus.DebugLevel)
		} else {
			logger.SetLevel(logrus.InfoLevel)
		}
	} else {
		logger.SetLevel(logrus.DebugLevel)
	}

	logger.Out = src

	// setting rotatelogs
	logWriter, err := rotatelogs.New(
		// Split file name
		fileName+".%Y%m%d.log",
		// Generate a soft chain and point to the latest log file
		rotatelogs.WithLinkName(fileName),
		// Set maximum save time (7 days)
		rotatelogs.WithMaxAge(7*24*time.Hour),
		// Set log cutting interval (1 day)
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
		TimestampFormat: "2006-01-02 15:04:05 +0800",
	}))

	// log debug
	// logger.WithFields().Info()
	// logger.WithFields(logrus.Fields{
	// 	"animal": "walrus",
	// }).Info("A walrus appears")

}

func GetLogger() *logrus.Logger {
	return logger
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
