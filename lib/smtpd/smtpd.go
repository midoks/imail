package smtpd

import (
	// "errors"
	"fmt"
	// "math/rand"
	"bufio"
	// "io/ioutil"
	"log"
	"net"
	"runtime"
	"strings"
	"time"
)

const (
	CMD_READY           = iota
	CMD_HELO            = iota
	CMD_AUTH_LOGIN      = iota
	CMD_AUTH_LOGIN_USER = iota
	CMD_AUTH_LOGIN_PWD  = iota
	CMD_MAIL_FROM       = iota
	CMD_RCPT_TO         = iota
	CMD_DATA            = iota
	CMD_QUIT            = iota
)

var stateList = map[int]string{
	CMD_HELO:       "HELO",
	CMD_AUTH_LOGIN: "AUTH LOGIN",
	CMD_MAIL_FROM:  "MAIL_FROM",
	CMD_RCPT_TO:    "RCPT_TO",
	CMD_DATA:       "DATA",
	CMD_QUIT:       "QUIT",
}

const (
	MSG_INIT            = "220.0"
	MSG_OK              = "220"
	MSG_BYE             = "221"
	MSG_BAD_SYNTAX      = "500"
	MSG_COMMAND_ERR     = "502"
	MSG_COMMAND_TM_ERR  = "421"
	MSG_AUTH_LOGIN_USER = "334.user"
	MSG_AUTH_LOGIN_PWD  = "334.passwd"
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
}

var GO_EOL = getGoEol()

func getGoEol() string {
	if "windows" == runtime.GOOS {
		return "\r\n"
	}
	return "\n"
}

type smtpdServer struct {
	conn      net.Conn
	state     int
	startTime time.Time
	errCount  int

	//save cmd info
	cmdHeloInfo string
}

func (this *smtpdServer) setState(state int) {
	this.state = state
}

func (this *smtpdServer) D(a ...interface{}) (n int, err error) {
	return fmt.Println(a...)
}

func (this *smtpdServer) write(code string) {

	info := fmt.Sprintf("%.3s %s%s", code, msgList[code], GO_EOL)
	_, err := this.conn.Write([]byte(info))

	if err != nil {
		log.Fatal(err)
	}
}

func (this *smtpdServer) getString0() (string, error) {

	input, err := bufio.NewReader(this.conn).ReadString('\n')
	if err != nil {
		return "", err
	}
	inputTrim := strings.TrimSpace(input)
	return inputTrim, err
}

func (this *smtpdServer) getString() (string, error) {
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

func (this *smtpdServer) close() {
	this.conn.Close()
}

func (this *smtpdServer) cmdCompare(input string, cmd int) bool {
	if strings.EqualFold(input, stateList[cmd]) {
		return true
	}
	return false
}

func (this *smtpdServer) cmdHelo(input string) bool {
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

func (this *smtpdServer) cmdAuthLogin(input string) bool {

	if this.cmdCompare(input, CMD_AUTH_LOGIN) {
		this.setState(CMD_AUTH_LOGIN)
		this.write(MSG_AUTH_LOGIN_USER)
		return true
	}
	this.write(MSG_BAD_SYNTAX)
	return false
}

func (this *smtpdServer) cmdAuthLoginUser(input string) bool {
	this.setState(CMD_AUTH_LOGIN_USER)
	this.write(MSG_AUTH_LOGIN_USER)
	return true
}

func (this *smtpdServer) cmdAuthLoginPwd(input string) bool {

	this.setState(CMD_AUTH_LOGIN_PWD)
	this.write(MSG_AUTH_LOGIN_USER)
	return true
}

func (this *smtpdServer) cmdMailFrom(input string) bool {

	if this.cmdCompare(input, CMD_MAIL_FROM) {
		this.setState(CMD_MAIL_FROM)
		this.write(MSG_AUTH_LOGIN_USER)
		return true
	}
	this.write(MSG_BAD_SYNTAX)
	return false
}

func (this *smtpdServer) cmdQuit(input string) bool {
	if this.cmdCompare(input, CMD_QUIT) {
		this.write(MSG_BYE)
		this.close()
		return true
	}
	return false
}

func (this *smtpdServer) handle() {

	for {
		state := this.state
		cmd, err := this.getString()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(state, cmd)

		if CMD_READY == state {

			if this.cmdHelo(cmd) {
				continue
			}

		} else if CMD_HELO == state {

			if this.cmdAuthLogin(cmd) {
				continue
			}

		} else if CMD_AUTH_LOGIN == state {
			if this.cmdAuthLoginUser(cmd) {
				continue
			}
		} else if CMD_AUTH_LOGIN_USER == state {
			if this.cmdAuthLoginPwd(cmd) {
				continue
			}
		} else if CMD_AUTH_LOGIN_PWD == state {

		} else if CMD_MAIL_FROM == state {

		} else if CMD_RCPT_TO == state {

		} else if CMD_DATA == state {

		}

		if this.cmdQuit(cmd) {
			break
		} else {
			this.write(MSG_COMMAND_ERR)
		}
	}
}

func (this *smtpdServer) start(conn net.Conn) {
	this.conn = conn
	this.startTime = time.Now()
	this.setState(CMD_READY)
	this.write(MSG_INIT)
	this.handle()
}

func Start() {

	ln, err := net.Listen("tcp", ":1025")
	if err != nil {
		panic(err)
		return
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()

		if err != nil {
			continue
		}
		srv := smtpdServer{}
		go srv.start(conn)
	}
}
