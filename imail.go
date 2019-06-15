package main

import (
	"fmt"
	"github.com/astaxie/beego/config"
	"github.com/midoks/imail/app"
	"github.com/midoks/imail/imap"
	"github.com/midoks/imail/pop3"
	"github.com/midoks/imail/smtpd"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/trace"
	"strconv"
)

func main() {

	go pprof()

	conf, err := config.NewConfig("ini", "conf/app.conf")
	if err != nil {
		fmt.Println("app config failed, err:", err)
		return
	}

	smptd_enable, err := conf.Bool("smtpd::enable")
	if err != nil {
		fmt.Println("read smptd:port failed, err:", err)
		return
		if smptd_enable {
			smptd_port, err := conf.Int("smtpd::port")
			if err != nil {
				fmt.Println("read smptd:port failed, err:", err)
				return
			}

			go smtpd.Start(smptd_port)
		}

	}

	pop3_enable, err := conf.Bool("pop3::enable")
	if err != nil {
		fmt.Println("read pop3:port failed, err:", err)
		return

		if pop3_enable {
			pop3_port, err := conf.Int("pop3::port")
			if err != nil {
				fmt.Println("read pop3:port failed, err:", err)
				return
			}
			go pop3.Start(pop3_port)
		}
	}

	imap_enable, err := conf.Bool("imap::enable")
	if err != nil {
		fmt.Println("read imap:port failed, err:", err)
		return

		if imap_enable {
			imap_port, err := conf.Int("imap::port")
			if err != nil {
				fmt.Println("read imap:port failed, err:", err)
				return
			}
			go imap.Start(imap_port)
		}
	}

	api_enable, err := conf.Bool("smtpd::enable")
	if err != nil {
		fmt.Println("read api:port enable, err:", err)

		if api_enable {
			api_port, err := conf.Int("api::port")
			if err != nil {
				fmt.Println("read pop3:port failed, err:", err)
				return
			}
			app.Start(api_port)
		}
	}

}

//手动GC
func gc(w http.ResponseWriter, r *http.Request) {
	runtime.GC()
	w.Write([]byte("StartGC"))
}

//运行trace
func traceStart(w http.ResponseWriter, r *http.Request) {
	f, err := os.Create("trace.out")
	if err != nil {
		panic(err)
	}

	err = trace.Start(f)
	if err != nil {
		panic(err)
	}
	w.Write([]byte("TrancStart"))
	fmt.Println("StartTrancs")
}

//停止trace
func traceStop(w http.ResponseWriter, r *http.Request) {
	trace.Stop()
	w.Write([]byte("TrancStop"))
	fmt.Println("StopTrancs")
}

// go tool trace trace.out

//运行pprof分析器
func pprof() {
	go func() {
		//关闭GC
		debug.SetGCPercent(-1)
		http.HandleFunc("/go_nums", func(w http.ResponseWriter, r *http.Request) {
			num := strconv.FormatInt(int64(runtime.NumGoroutine()), 10)
			w.Write([]byte(num))
		})
		//运行trace
		http.HandleFunc("/start", traceStart)
		//停止trace
		http.HandleFunc("/stop", traceStop)
		//手动GC
		http.HandleFunc("/gc", gc)
		//网站开始监听
		http.ListenAndServe(":6060", nil)
	}()
}
