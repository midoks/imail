package debug

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/trace"
	"strconv"
)

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
func Pprof() {
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
