package main

import (
	"fmt"
	"github.com/midoks/imail/internal/app"
	"github.com/midoks/imail/internal/config"
	"github.com/midoks/imail/internal/db"
	// "github.com/midoks/imail/internal/dkim"
	"github.com/fsnotify/fsnotify"
	"github.com/midoks/imail/internal/debug"
	"github.com/midoks/imail/internal/imap"
	"github.com/midoks/imail/internal/log"
	"github.com/midoks/imail/internal/pop3"
	"github.com/midoks/imail/internal/smtpd"
	"strings"
)

func startService(name string) {
	config_enable := fmt.Sprintf("%s.enable", name)
	enable, err := config.GetBool(config_enable, false)
	if err == nil && enable {

		config_port := fmt.Sprintf("%s.port", name)
		port, err := config.GetInt(config_port, 25)
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
	ssl_enable, err := config.GetBool(config_ssl_enable, false)
	if err == nil && ssl_enable {

		config_ssl_port := fmt.Sprintf("%s.ssl_port", name)
		ssl_port, err := config.GetInt(config_ssl_port, 25)
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

func StartMonitor(path string) {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("StartMonitor:err", err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case e := <-watcher.Events:

				if e.Op&fsnotify.Chmod == fsnotify.Chmod {
					fmt.Printf("%s had change content!", path)
				}

			case err = <-watcher.Errors:
				if err != nil {
					fmt.Println("错误:", err)
				}

			}
		}
	}()

	err = watcher.Add(path)
	if err != nil {
		fmt.Println("Failed to watch directory: ", err)
	}
	<-done
}

func main() {
	// go mod init
	// go mod tidy
	// go mod vendor
	log.Init()

	err := config.Load("conf/app.conf")
	if err != nil {
		log.Panicf("config file load err")
	}

	db.Init()

	runmode := config.GetString("runmode", "dev")
	if strings.EqualFold(runmode, "dev") {
		go debug.Pprof()
	}

	startService("smtpd")
	startService("pop3")
	startService("imap")

	http_enable, err := config.GetBool("http.enable", false)
	if http_enable {
		http_port, err := config.GetInt("http.port", 80)
		if err == nil {
			log.Info("listen http success!")
			app.Start(http_port)
		} else {
			log.Errorf("listen http erorr:%s", err)
		}
	}
}
