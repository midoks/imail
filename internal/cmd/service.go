package cmd

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/midoks/imail/internal/app"
	"github.com/midoks/imail/internal/conf"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/imap"
	"github.com/midoks/imail/internal/log"
	"github.com/midoks/imail/internal/pop3"
	"github.com/midoks/imail/internal/smtpd"
	"github.com/midoks/imail/internal/task"
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

func newService(confFile string) {

	logger := log.Init()

	format := conf.GetString("log.format", "json")
	if strings.EqualFold(format, "json") {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else if strings.EqualFold(format, "text") {
		logger.SetFormatter(&logrus.TextFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{})
	}

	runmode := conf.GetString("runmode", "dev")
	if strings.EqualFold(runmode, "dev") {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	go ConfigFileStartMonitor(confFile)

	err := db.Init()
	if err != nil {
		return
	}

	if strings.EqualFold(runmode, "dev") {
		go debug.Pprof()
	}

	task.Init()

	startService("smtpd")
	startService("pop3")
	startService("imap")

	http_enable, err := conf.GetBool("web.enable", false)
	// fmt.Println(http_enable)
	fmt.Println(http_enable)
	if http_enable {
		http_port, err := conf.GetInt("web.port", 80)
		fmt.Println(http_port)
		if err == nil {
			log.Infof("listen http[%d] success!", http_port)
			app.Start(http_port)
		} else {
			log.Errorf("listen http[%d] erorr:%s", http_port, err)
		}
	}
}

func runAllService(c *cli.Context) error {

	confFile, err := initConfig(c, "")
	if err != nil {
		panic("imail config file load error")
		return err
	}

	fmt.Println(confFile)

	newService(confFile)
	return nil
}

func ServiceDebug() {

	confFile, err := initConfig(nil, "conf/app.conf")
	if err != nil {
		panic("imail config file load error")
	}

	newService(confFile)
}

func startService(name string) {
	config_enable := fmt.Sprintf("%s.enable", name)
	enable, err := conf.GetBool(config_enable, false)
	if err == nil && enable {

		config_port := fmt.Sprintf("%s.port", name)
		port, err := conf.GetInt(config_port, 25)
		if err == nil {
			log.Infof("listen %s port:%d success!", name, port)

			if strings.EqualFold(name, "smtpd") {
				go smtpd.Start(port)
			} else if strings.EqualFold(name, "pop3") {
				go pop3.Start(port)
			} else if strings.EqualFold(name, "imap") {
				go imap.Start(port)
			}
		} else {
			log.Errorf("listen %s erorr:%s", name, err)
		}
	}

	config_ssl_enable := fmt.Sprintf("%s.ssl_enable", name)
	ssl_enable, err := conf.GetBool(config_ssl_enable, false)
	if err == nil && ssl_enable {

		config_ssl_port := fmt.Sprintf("%s.ssl_port", name)
		ssl_port, err := conf.GetInt(config_ssl_port, 25)
		if err == nil {
			log.Infof("listen %s ssl port:%d success!", name, ssl_port)

			if strings.EqualFold(name, "smtpd") {
				go smtpd.StartSSL(ssl_port)
			} else if strings.EqualFold(name, "pop3") {
				go pop3.StartSSL(ssl_port)
			} else if strings.EqualFold(name, "imap") {
				go imap.StartSSL(ssl_port)
			}
		} else {
			log.Errorf("listen %s ssl erorr:%s", name, err)
		}
	}
}

func reloadService(path string) {
	log.Infof("reloadService:%s", path)

	err := conf.Load(path)
	if err != nil {
		log.Errorf("imail config file reload error:%s", err)
		return
	}

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
