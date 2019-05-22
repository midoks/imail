package smtpd

import (
	// "errors"
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/trace"
	"strings"
	"time"
)

const (
	CMD_READY           = iota
	CMD_HELO            = iota
	CMD_EHLO            = iota
	CMD_AUTH_LOGIN      = iota
	CMD_AUTH_LOGIN_USER = iota
	CMD_AUTH_LOGIN_PWD  = iota
	CMD_MAIL_FROM       = iota
	CMD_RCPT_TO         = iota
	CMD_DATA            = iota
	CMD_DATA_END        = iota
	CMD_QUIT            = iota
)

var mailFlow = []int{
	CMD_READY,
	CMD_HELO,
	CMD_AUTH_LOGIN,
	CMD_AUTH_LOGIN_USER,
	CMD_AUTH_LOGIN_PWD,
	CMD_MAIL_FROM,
	CMD_RCPT_TO,
	CMD_DATA,
	CMD_DATA_END,
	CMD_QUIT,
}

var stateList = map[int]string{
	CMD_READY:      "READY",
	CMD_HELO:       "HELO",
	CMD_EHLO:       "EHLO",
	CMD_AUTH_LOGIN: "AUTH LOGIN",
	CMD_MAIL_FROM:  "MAIL FROM",
	CMD_RCPT_TO:    "RCPT TO",
	CMD_DATA:       "DATA",
	CMD_DATA_END:   ".",
	CMD_QUIT:       "QUIT",
}

const (
	MSG_INIT            = "220.init"
	MSG_OK              = "220"
	MSG_MAIL_OK         = "250"
	MSG_BYE             = "221"
	MSG_BAD_SYNTAX      = "500"
	MSG_COMMAND_ERR     = "502"
	MSG_COMMAND_TM_ERR  = "421"
	MSG_AUTH_LOGIN_USER = "334.user"
	MSG_AUTH_LOGIN_PWD  = "334.passwd"
	MSG_AUTH_OK         = "235"
	MSG_AUTH_FAIL       = "535"
	MSG_DATA            = "354"
)

var msgList = map[string]string{
	MSG_INIT:            "Anti-spam GT for Coremail System(imail)",
	MSG_OK:              "ok",
	MSG_BYE:             "bye",
	MSG_COMMAND_ERR:     "Error: command not implemented",
	MSG_COMMAND_TM_ERR:  "Too many error commands",
	MSG_BAD_SYNTAX:      "Error: bad syntax",
	MSG_AUTH_LOGIN_USER: "dXNlcm5hbWU6",
	MSG_AUTH_LOGIN_PWD:  "UGFzc3dvcmQ6",
	MSG_AUTH_OK:         "Authentication successful",
	MSG_AUTH_FAIL:       "Error: authentication failed",
	MSG_MAIL_OK:         "Mail OK",
	MSG_DATA:            "End data with <CR><LF>.<CR><LF>",
}

var GO_EOL = GetGoEol()

func GetGoEol() string {
	if "windows" == runtime.GOOS {
		return "\r\n"
	}
	return "\n"
}

type SmtpService struct {
	*FSM
}

func NewSmtpd(initState FSMState) *SmtpService {
	return &SmtpService{
		FSM: NewFSM(initState),
	}
}

type SmtpdServer struct {
	debug       bool
	conn        net.Conn
	state       int
	startTime   time.Time
	errCount    int
	srv         *SmtpService
	cmdHeloInfo string
}

func (this *SmtpdServer) setState(state int) {
	this.state = state
}

func (this *SmtpdServer) D(a ...interface{}) (n int, err error) {
	return fmt.Println(a...)
}

func (this *SmtpdServer) Debug(d bool) {
	this.debug = d
}

func (this *SmtpdServer) write(code string) {

	info := fmt.Sprintf("%.3s %s%s", code, msgList[code], GO_EOL)
	_, err := this.conn.Write([]byte(info))

	if err != nil {
		log.Fatal(err)
	}
}

func (this *SmtpdServer) getString() (string, error) {
	input, err := bufio.NewReader(this.conn).ReadString('\n')
	if err != nil {
		return "", err
	}
	inputTrim := strings.TrimSpace(input)
	return inputTrim, err
}

func (this *SmtpdServer) getString0() (string, error) {
	buffer := make([]byte, 2048)

	n, err := this.conn.Read(buffer)
	if err != nil {
		log.Fatal(this.conn.RemoteAddr().String(), " connection error: ", err)
		return "", err
	}

	input := string(buffer[:n])
	inputTrim := strings.TrimSpace(input)
	return inputTrim, err
}

func (this *SmtpdServer) close() {
	this.conn.Close()
}

func (this *SmtpdServer) cmdCompare(input string, cmd int) bool {
	fmt.Println(input, stateList[cmd])
	if strings.EqualFold(input, stateList[cmd]) {
		return true
	}
	return false
}

func (this *SmtpdServer) cmdHelo(input string) bool {
	inputN := strings.SplitN(input, " ", 2)

	if this.cmdCompare(inputN[0], CMD_HELO) {
		if len(inputN) < 2 {
			this.write(MSG_BAD_SYNTAX)
			return false
		}

		this.setState(CMD_HELO)
		this.write(MSG_OK)
		return true
	}
	this.write(MSG_COMMAND_ERR)
	return false
}

func (this *SmtpdServer) cmdAuthLogin(input string) bool {

	if this.cmdCompare(input, CMD_AUTH_LOGIN) {
		this.setState(CMD_AUTH_LOGIN)
		this.write(MSG_AUTH_LOGIN_USER)
		return true
	}
	this.write(MSG_BAD_SYNTAX)
	return false
}

func (this *SmtpdServer) cmdAuthLoginUser(input string) bool {
	this.setState(CMD_AUTH_LOGIN_USER)
	this.write(MSG_AUTH_LOGIN_PWD)
	return true
}

func (this *SmtpdServer) cmdAuthLoginPwd(input string) bool {

	this.setState(CMD_AUTH_LOGIN_PWD)
	this.write(MSG_AUTH_OK)
	return true
}

func (this *SmtpdServer) cmdMailFrom(input string) bool {
	inputN := strings.SplitN(input, ":", 2)
	if this.cmdCompare(inputN[0], CMD_MAIL_FROM) {
		this.setState(CMD_MAIL_FROM)
		this.write(MSG_MAIL_OK)
		return true
	}
	this.write(MSG_BAD_SYNTAX)
	return false
}

func (this *SmtpdServer) cmdRcptTo(input string) bool {
	inputN := strings.SplitN(input, ":", 2)
	if this.cmdCompare(inputN[0], CMD_RCPT_TO) {
		this.setState(CMD_RCPT_TO)
		this.write(MSG_MAIL_OK)
		return true
	}
	this.write(MSG_BAD_SYNTAX)
	return false
}

func (this *SmtpdServer) cmdData(input string) bool {
	if this.cmdCompare(input, CMD_DATA) {
		this.setState(CMD_DATA)
		this.write(MSG_DATA)
		return true
	}
	this.write(MSG_BAD_SYNTAX)
	return false
}

func (this *SmtpdServer) cmdDataEnd(input string) bool {
	if this.cmdCompare(input, CMD_DATA_END) {
		this.setState(CMD_DATA_END)
		this.write(MSG_DATA)
		return true
	}
	this.write(MSG_BAD_SYNTAX)
	return false
}

func (this *SmtpdServer) cmdQuit(input string) bool {
	if this.cmdCompare(input, CMD_QUIT) {
		this.write(MSG_BYE)
		this.close()
		return true
	}
	return false
}

func (this *SmtpdServer) cmdCommon(input string) bool {

	inputN := strings.SplitN(input, " ", 2)
	// state := this.state

	if this.cmdCompare(inputN[0], CMD_HELO) {

		this.setState(CMD_HELO)
		this.write(MSG_OK)
		return true
	}
	this.write(MSG_COMMAND_ERR)
	return false
}

func (this *SmtpdServer) Call(input string) {
	this.cmdCommon(input)
}

func (this *SmtpdServer) handle() {

	// for {

	state := string(this.srv.getState())

	if strings.EqualFold(state, stateList[CMD_READY]) {
		this.srv.Call(CMD_READY_FE)
	}

	if strings.EqualFold(state, stateList[CMD_HELO]) {
		this.srv.Call(CMD_HELO_FE)
	}

	if strings.EqualFold(state, stateList[CMD_EHLO]) {
		this.srv.Call(CMD_EHLO_FE)
	}
	// }
}

var (
	CMD_READY_FS = FSMState(stateList[CMD_READY])
	CMD_READY_FE = FSMEvent(stateList[CMD_READY])

	CMD_HELO_FS = FSMState(stateList[CMD_HELO])
	CMD_HELO_FE = FSMEvent(stateList[CMD_HELO])

	CMD_EHLO_FS = FSMState(stateList[CMD_EHLO])
	CMD_EHLO_FE = FSMEvent(stateList[CMD_EHLO])
)

func (this *SmtpdServer) register() {
	var (
		CMD_READY_FH = FSMHandler(func() FSMState {

			input, _ := this.getString()

			this.cmdQuit(input)

			if strings.EqualFold(input, stateList[CMD_HELO]) {
				return FSMState(stateList[CMD_HELO])
			} else if strings.EqualFold(input, stateList[CMD_EHLO]) {
				return FSMState(stateList[CMD_EHLO])
			}

			return FSMState(stateList[CMD_READY])
		})

		CMD_HELO_FH = FSMHandler(func() FSMState {
			input, _ := this.getString()

			fmt.Println(input)
			return FSMState("CMD_HELO")
		})

		CMD_EHLO_FH = FSMHandler(func() FSMState {
			input, _ := this.getString()

			fmt.Println(input)
			return FSMState("CMD_HELOs")
		})
	)

	this.srv = NewSmtpd(CMD_READY_FS)
	this.srv.AddHandler(CMD_READY_FS, CMD_READY_FE, CMD_READY_FH)
	this.srv.AddHandler(CMD_READY_FS, CMD_HELO_FE, CMD_HELO_FH)
	this.srv.AddHandler(CMD_READY_FS, CMD_EHLO_FE, CMD_EHLO_FH)
}

func (this *SmtpdServer) start(conn net.Conn) {
	this.conn = conn
	conn.SetReadDeadline(time.Now().Add(time.Minute * 180))
	defer conn.Close()

	this.startTime = time.Now()

	this.write(MSG_INIT)
	this.setState(CMD_READY)

	this.register()
	this.handle()
}

func Start() {

	ln, err := net.Listen("tcp", ":1025")
	defer ln.Close()
	if err != nil {
		panic(err)
		return
	}
	defer ln.Close()

	go pprof()

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		srv := SmtpdServer{}
		go srv.start(conn)
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
