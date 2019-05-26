package main

import (
	"fmt"
	"github.com/astaxie/beego/config"
	"github.com/midoks/imail/app"
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

	api_enable, err := conf.Bool("smtpd::enable")
	if err != nil {
		fmt.Println("read api:port enable, err:", err)
		return
	}

	if api_enable {
		app.Start()
	}
}
