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

type SmtpdServer struct {
	debug     bool
	conn      net.Conn
	state     int
	startTime time.Time
	errCount  int
	// srv         *SmtpService
	cmdHeloInfo string
}

func (this *SmtpdServer) setState(state int) {
	this.state = state
}

func (this *SmtpdServer) getState() int {
	return this.state
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
	if strings.EqualFold(input, stateList[cmd]) {
		return true
	}
	return false
}

func (this *SmtpdServer) stateCompare(input int, cmd int) bool {
	if input == cmd {
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

func (this *SmtpdServer) cmdEhlo(input string) bool {
	inputN := strings.SplitN(input, " ", 2)

	if this.cmdCompare(inputN[0], CMD_EHLO) {
		if len(inputN) < 2 {
			this.write(MSG_BAD_SYNTAX)
			return false
		}

		this.setState(CMD_EHLO)
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

	if this.cmdCompare(inputN[0], CMD_HELO) {

		this.setState(CMD_HELO)
		this.write(MSG_OK)
		return true
	}

	this.write(MSG_COMMAND_ERR)
	return false
}

func (this *SmtpdServer) handle() {
	for {
		state := this.getState()
		input, _ := this.getString()

		switch state {
		case CMD_READY:
			if this.cmdQuit(input) {
				break
			}

			if this.cmdHelo(input) {
				this.setState(CMD_HELO)
			} else if this.cmdEhlo(input) {
				this.setState(CMD_EHLO)
			} else {
				this.write(MSG_COMMAND_ERR)
			}
		case CMD_HELO:
			if this.cmdQuit(input) {
				break
			}

			if this.cmdAuthLogin(input) {
				this.setState(CMD_AUTH_LOGIN)
			}
		case CMD_EHLO:
			if this.cmdQuit(input) {
			}

			if this.cmdMailFrom(input) {
				this.setState(CMD_MAIL_FROM)
			}
		case CMD_AUTH_LOGIN:
			if this.cmdAuthLoginUser(input) {
				this.setState(CMD_AUTH_LOGIN_USER)
			}
		case CMD_AUTH_LOGIN_USER:
			if this.cmdAuthLoginPwd(input) {
				this.setState(CMD_AUTH_LOGIN_PWD)
			}
		case CMD_AUTH_LOGIN_PWD:
			if this.cmdMailFrom(input) {
				this.setState(CMD_MAIL_FROM)
			}
		case CMD_MAIL_FROM:
			if this.cmdQuit(input) {
				break
			}
			if this.cmdRcptTo(input) {
				this.setState(CMD_RCPT_TO)
			}

		case CMD_RCPT_TO:
			if this.cmdQuit(input) {
				break
			}
			if this.cmdRcptTo(input) {
				this.setState(CMD_DATA)
			}
		case CMD_DATA:
			if this.cmdData(input) {
				this.setState(CMD_DATA_END)
			}

		case CMD_DATA_END:
			if this.cmdQuit(input) {
				break
			}
			if this.cmdDataEnd(input) {
				this.setState(CMD_READY)
			}
		}
	}
}

func (this *SmtpdServer) start(conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(time.Minute * 3))
	defer conn.Close()
	this.conn = conn

	this.startTime = time.Now()

	this.write(MSG_INIT)
	this.setState(CMD_READY)

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
