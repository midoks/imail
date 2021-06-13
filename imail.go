package main

import (
	"fmt"
	"github.com/midoks/imail/internal/app"
	"github.com/midoks/imail/internal/config"
	"github.com/midoks/imail/internal/db"
	// "github.com/midoks/imail/internal/dkim"
	ipserver "github.com/midoks/imail/internal/imap"
	"github.com/midoks/imail/internal/pop3"
	"github.com/midoks/imail/internal/smtpd"

	// "io/ioutil"
	// "net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/trace"
	"strconv"
	// "strings"
)

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

	go pprof()

	db.Init()

	smptd_enable := config.GetBool("smtpd.enable")

	if smptd_enable {
		smptd_port := config.GetInt("smtpd.port")

		go smtpd.Start(smptd_port)
		fmt.Println("listen smtpd success!")
	}

	pop3_enable := config.GetBool("pop3.enable")

	if pop3_enable {
		pop3_port := config.GetInt("pop3.port")

		go pop3.Start(pop3_port)
		fmt.Println("listen pop3 success!")
	}

	imap_enable := config.GetBool("imap.enable")
	if imap_enable {
		imap_port := config.GetInt("imap.port")

		go ipserver.Start(imap_port)
		fmt.Println("listen imap success!")
	}

	http_enable := config.GetBool("http.enable")
	if http_enable {
		http_port := config.GetInt("http.port")

		app.Start(http_port)
		fmt.Println("listen http success!")
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
