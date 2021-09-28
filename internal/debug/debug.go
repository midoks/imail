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

//start trace
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

//stop trace
func traceStop(w http.ResponseWriter, r *http.Request) {
	trace.Stop()
	w.Write([]byte("TrancStop"))
	fmt.Println("StopTrancs")
}

// go tool trace trace.out
// Run pprof analyzer
// http://127.0.0.1:6060/debug/pprof/

// code a 1:
// http://localhost:6060/debug/pprof/profile?seconds=30
// go tool pprof -http=:8080 profile
func Pprof() {
	go func() {
		//Close GC

		debug.SetGCPercent(-1)

		http.HandleFunc("/go_nums", func(w http.ResponseWriter, r *http.Request) {
			num := strconv.FormatInt(int64(runtime.NumGoroutine()), 10)
			w.Write([]byte(num))
		})

		http.HandleFunc("/start", traceStart)
		http.HandleFunc("/stop", traceStop)
		http.HandleFunc("/gc", gc)
		http.ListenAndServe(":6060", nil)
	}()
}
