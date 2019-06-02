package main

import (
	"fmt"
	"github.com/astaxie/beego/config"
	"github.com/midoks/imail/app"
	"github.com/midoks/imail/pop3"
	"github.com/midoks/imail/smtpd"
)

func main() {

	conf, err := config.NewConfig("ini", "conf/app.conf")
	if err != nil {
		fmt.Println("app config failed, err:", err)
		return
	}

	smptd_enable, err := conf.Bool("smtpd::enable")
	if err != nil {
		fmt.Println("read smptd:port failed, err:", err)
		return
	}

	if smptd_enable {
		smptd_port, err := conf.Int("smtpd::port")
		if err != nil {
			fmt.Println("read smptd:port failed, err:", err)
			return
		}

		go smtpd.Start(smptd_port)
	}

	pop3_enable, err := conf.Bool("pop3::enable")
	if err != nil {
		fmt.Println("read pop3:port failed, err:", err)
		return
	}

	if pop3_enable {
		pop3_port, err := conf.Int("pop3::port")
		if err != nil {
			fmt.Println("read pop3:port failed, err:", err)
			return
		}
		go pop3.Start(pop3_port)
	}

	api_enable, err := conf.Bool("smtpd::enable")
	if err != nil {
		fmt.Println("read api:port enable, err:", err)
		return
	}

	if api_enable {
		app.Start()
	}
}
