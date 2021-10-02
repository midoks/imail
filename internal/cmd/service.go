package cmd

import (
	// "fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/midoks/imail/internal/app"
	"github.com/midoks/imail/internal/app/router"
	"github.com/midoks/imail/internal/conf"
	// "github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/imap"
	"github.com/midoks/imail/internal/log"
	"github.com/midoks/imail/internal/pop3"
	"github.com/midoks/imail/internal/smtpd"
	// "github.com/midoks/imail/internal/task"
	"github.com/midoks/imail/internal/tools/debug"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"strings"
)

var Service = cli.Command{
	Name:        "service",
	Usage:       "This command starts all services",
	Description: `Start POP3, IMAP, SMTP, web and other services`,
	Action:      runAllService,
	Flags: []cli.Flag{
		stringFlag("config, c", "", "Custom configuration file path"),
	},
}

func newService() {

	logger := log.Init()

	format := conf.Log.Format
	if strings.EqualFold(format, "json") {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else if strings.EqualFold(format, "text") {
		logger.SetFormatter(&logrus.TextFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{})
	}

	if strings.EqualFold(conf.App.RunMode, "dev") {
		logger.SetLevel(logrus.DebugLevel)
		go debug.Pprof()
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	// err := db.Init()
	// if err != nil {
	// 	return
	// }

	// task.Init()

	// startService("smtpd")
	// startService("pop3")
	// startService("imap")

	app.Start(conf.Web.HttpPort)

}

func runAllService(c *cli.Context) error {

	err := router.GlobalInit(c.String("config"))
	if err != nil {
		log.Fatal("Failed to initialize application: %v", err)
	}

	newService()
	return nil
}

func ServiceDebug() {

	err := router.GlobalInit("")
	if err != nil {
		log.Fatal("Failed to initialize application: %v", err)
	}
	newService()
}

func startService(name string) {

	if strings.EqualFold(name, "smtpd") && conf.Smtp.Enable {
		go smtpd.Start(conf.Smtp.Port)
	} else if strings.EqualFold(name, "pop3") && conf.Pop3.Enable {
		go pop3.Start(conf.Pop3.Port)
	} else if strings.EqualFold(name, "imap") && conf.Imap.Enable {
		go imap.Start(conf.Imap.Port)
	}

	log.Infof("listen %s success!", name)

	if strings.EqualFold(name, "smtpd") && conf.Smtp.SslEnable {
		go smtpd.StartSSL(conf.Smtp.Port)
	} else if strings.EqualFold(name, "pop3") && conf.Pop3.SslEnable {
		go pop3.StartSSL(conf.Pop3.Port)
	} else if strings.EqualFold(name, "imap") && conf.Imap.SslEnable {
		go imap.StartSSL(conf.Imap.Port)
	}

	log.Infof("listen %s ssl success!", name)

}

func reloadService(path string) {
	log.Infof("reloadService:%s", path)

	// fmt.Println("imap reload start")
	// err = imap.Close()
	// fmt.Println("[reloadService]err", err)
	// fmt.Println("startService")
	// startService("imap")
	// fmt.Println("imap reload end")

}

func ConfigFileStartMonitor(path string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Errorf("fsnotify.NewWatcher errors:%s", err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case e := <-watcher.Events:

				if e.Op&fsnotify.Chmod == fsnotify.Chmod {
					reloadService(path)
				}
			case err = <-watcher.Errors:
				if err != nil {
					log.Errorf("watcher errors:%s", err)
				}
			}
		}
	}()

	err = watcher.Add(path)
	if err != nil {
		log.Errorf("failed to watch directory error:%s", err)
	}
	<-done
}
