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
	"github.com/midoks/imail/internal/pop3"
	"github.com/midoks/imail/internal/smtpd"
	// "log"
	"strings"
)

func startService(name string) {
	config_enable := fmt.Sprintf("%s.enable", name)
	enable, err := config.GetBool(config_enable, false)
	if err == nil && enable {

		config_port := fmt.Sprintf("%s.port", name)
		port, err := config.GetInt(config_port, 25)
		if err == nil {
			fmt.Printf("listen %s port:%d success!\n", name, port)

			if strings.EqualFold(name, "smtpd") {
				go smtpd.Start(port)
			} else if strings.EqualFold(name, "pop3") {
				go pop3.Start(port)
			} else if strings.EqualFold(name, "imap") {
				go imap.Start(port)
			}
		} else {
			fmt.Printf("listen %s erorr:%s\n", name, err)
		}
	}

	config_ssl_enable := fmt.Sprintf("%s.ssl_enable", name)
	ssl_enable, err := config.GetBool(config_ssl_enable, false)
	if err == nil && ssl_enable {

		config_ssl_port := fmt.Sprintf("%s.ssl_port", name)
		ssl_port, err := config.GetInt(config_ssl_port, 25)
		if err == nil {
			fmt.Printf("listen %s ssl port:%d success!\n", name, ssl_port)

			if strings.EqualFold(name, "smtpd") {
				go smtpd.StartSSL(ssl_port)
			} else if strings.EqualFold(name, "pop3") {
				go pop3.StartSSL(ssl_port)
			} else if strings.EqualFold(name, "imap") {
				go imap.StartSSL(ssl_port)
			}
		} else {
			fmt.Printf("listen %s ssl erorr:%s\n", name, err)
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

	err := config.Load("conf/app.conf")
	if err != nil {
		return
	}
	// err := dkim.MakeDkimConfFile("biqu.xyz")
	// fmt.Println(err)

	// tomail := "627293072@qq.com"

	// msg := []byte("from:admin@cachecha.com\r\n" +
	// 	"to: " + tomail + "\r\n" +
	// 	"Subject: hello,imail!\r\n" +
	// 	"Content-Type:multipart/mixed;boundary=a\r\n" +
	// 	"Mime-Version:1.0\r\n" +
	// 	"\r\n" +
	// 	"--a\r\n" +
	// 	"Content-type:text/plain;charset=utf-8\r\n" +
	// 	"Content-Transfer-Encoding:quoted-printable\r\n" +
	// 	"\r\n" +
	// 	"此处为正文内容D!\r\n")

	// err := smtpd.Delivery("admin@cachecha.com", tomail, msg)
	// fmt.Println("err:", err)

	// auth := smtpd.PlainAuth("", "yuludejia@gmail.com", "pmroenyllybhlwub", "smtp.gmail.com")

	// msg := []byte("from:yuludejia@gmail.com\r\n" +
	// 	"to: midoks@163.com\r\n" +
	// 	"Subject: hello,subject!\r\n" +
	// 	"Content-Type:multipart/mixed;boundary=a\r\n" +
	// 	"Mime-Version:1.0\r\n" +
	// 	"\r\n" +
	// 	"--a\r\n" +
	// 	"Content-type:text/plain;charset=utf-8\r\n" +
	// 	"Content-Transfer-Encoding:quoted-printable\r\n" +
	// 	"\r\n" +
	// 	"此处为正文内容D!\r\n")

	// err := smtpd.SendMail("smtp.gmail.com", "587", auth, "yuludejia@gmail.com", []string{"midoks@163.com"}, msg)
	// fmt.Println("err:", err)

	runmode := config.GetString("runmode", "dev")
	if strings.EqualFold(runmode, "dev") {
		go debug.Pprof()
	}

	db.Init()

	startService("smtpd")
	startService("pop3")
	startService("imap")

	http_enable, err := config.GetBool("http.enable", false)
	if http_enable {
		http_port, err := config.GetInt("http.port", 80)
		if err == nil {
			app.Start(http_port)
			fmt.Println("listen http success!")
		} else {
			fmt.Println("listen http erorr:", err)
		}

	}
	fmt.Println("end", err)
}
