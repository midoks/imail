package admin

import (
	"fmt"
	"runtime"
	// "strings"
	"time"

	_ "github.com/json-iterator/go"

	"github.com/midoks/imail/internal/app/context"
	"github.com/midoks/imail/internal/conf"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/task"
	"github.com/midoks/imail/internal/tools"
)

const (
	tmplDashboard = "admin/dashboard"
	tmplConfig    = "admin/config"
	tmplMonitor   = "admin/monitor"
)

// initTime is the time when the application was initialized.
var initTime = time.Now()

var sysStatus struct {
	Uptime       string
	NumGoroutine int

	// General statistics.
	MemAllocated string // bytes allocated and still in use
	MemTotal     string // bytes allocated (even if freed)
	MemSys       string // bytes obtained from system (sum of XxxSys below)
	Lookups      uint64 // number of pointer lookups
	MemMallocs   uint64 // number of mallocs
	MemFrees     uint64 // number of frees

	// Main allocation heap statistics.
	HeapAlloc    string // bytes allocated and still in use
	HeapSys      string // bytes obtained from system
	HeapIdle     string // bytes in idle spans
	HeapInuse    string // bytes in non-idle span
	HeapReleased string // bytes released to the OS
	HeapObjects  uint64 // total number of allocated objects

	// Low-level fixed-size structure allocator statistics.
	//	Inuse is bytes used now.
	//	Sys is bytes obtained from system.
	StackInuse  string // bootstrap stacks
	StackSys    string
	MSpanInuse  string // mspan structures
	MSpanSys    string
	MCacheInuse string // mcache structures
	MCacheSys   string
	BuckHashSys string // profiling bucket hash table
	GCSys       string // GC metadata
	OtherSys    string // other system allocations

	// Garbage collector statistics.
	NextGC       string // next run in HeapAlloc time (bytes)
	LastGC       string // last run in absolute time (ns)
	PauseTotalNs string
	PauseNs      string // circular buffer of recent GC pause times, most recent at [(NumGC+255)%256]
	NumGC        uint32
}
var m *runtime.MemStats

func init() {
	// go func() {
	// 	updateSystemStatus()
	// 	time.Sleep(time.Microsecond * 100)
	// }()

	m = new(runtime.MemStats)
}

func updateSystemStatus() {

	runtime.ReadMemStats(m)

	sysStatus.Uptime = tools.TimeSincePro(initTime)

	sysStatus.NumGoroutine = runtime.NumGoroutine()

	sysStatus.MemAllocated = tools.ToSize(int64(m.Alloc))
	sysStatus.MemTotal = tools.ToSize(int64(m.TotalAlloc))
	sysStatus.MemSys = tools.ToSize(int64(m.Sys))
	sysStatus.Lookups = m.Lookups
	sysStatus.MemMallocs = m.Mallocs
	sysStatus.MemFrees = m.Frees

	sysStatus.HeapAlloc = tools.ToSize(int64(m.HeapAlloc))
	sysStatus.HeapSys = tools.ToSize(int64(m.HeapSys))
	sysStatus.HeapIdle = tools.ToSize(int64(m.HeapIdle))
	sysStatus.HeapInuse = tools.ToSize(int64(m.HeapInuse))
	sysStatus.HeapReleased = tools.ToSize(int64(m.HeapReleased))
	sysStatus.HeapObjects = m.HeapObjects

	sysStatus.StackInuse = tools.ToSize(int64(m.StackInuse))
	sysStatus.StackSys = tools.ToSize(int64(m.StackSys))
	sysStatus.MSpanInuse = tools.ToSize(int64(m.MSpanInuse))
	sysStatus.MSpanSys = tools.ToSize(int64(m.MSpanSys))
	sysStatus.MCacheInuse = tools.ToSize(int64(m.MCacheInuse))
	sysStatus.MCacheSys = tools.ToSize(int64(m.MCacheSys))
	sysStatus.BuckHashSys = tools.ToSize(int64(m.BuckHashSys))
	sysStatus.GCSys = tools.ToSize(int64(m.GCSys))
	sysStatus.OtherSys = tools.ToSize(int64(m.OtherSys))

	sysStatus.NextGC = tools.ToSize(int64(m.NextGC))
	sysStatus.LastGC = fmt.Sprintf("%.1fs", float64(time.Now().UnixNano()-int64(m.LastGC))/1000/1000/1000)
	sysStatus.PauseTotalNs = fmt.Sprintf("%.1fs", float64(m.PauseTotalNs)/1000/1000/1000)
	sysStatus.PauseNs = fmt.Sprintf("%.3fs", float64(m.PauseNs[(m.NumGC+255)%256])/1000/1000/1000)
	sysStatus.NumGC = m.NumGC
}

func Dashboard(c *context.Context) {
	c.Title("admin.dashboard")
	c.PageIs("Admin")
	c.PageIs("AdminDashboard")

	c.Data["GoVersion"] = runtime.Version()
	c.Data["BuildTime"] = conf.BuildTime
	c.Data["BuildCommit"] = conf.BuildCommit

	c.Data["Stats"] = db.GetStatistic()
	// // FIXME: update periodically

	updateSystemStatus()
	c.Data["SysStatus"] = sysStatus
	c.Success(tmplDashboard)
}

func Monitor(c *context.Context) {
	c.Data["Title"] = c.Tr("admin.monitor")
	c.Data["PageIsAdmin"] = true
	c.Data["PageIsAdminMonitor"] = true
	// c.Data["Processes"] = process.Processes
	c.Data["Entries"] = task.ListTasks()
	c.Success(tmplMonitor)
}

func Config(c *context.Context) {
	c.Title("admin.config")
	c.PageIs("Admin")
	c.PageIs("AdminConfig")

	c.Data["App"] = conf.App
	c.Data["Web"] = conf.Web
	c.Data["Database"] = conf.Database
	c.Data["Security"] = conf.Security
	c.Data["Session"] = conf.Session
	c.Data["Cache"] = conf.Cache
	c.Data["LogRootPath"] = conf.Log.RootPath

	c.Success(tmplConfig)
}
