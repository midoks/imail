package pop3

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	_ "net/http/pprof"
	"runtime"
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

type Pop3Server struct {
	debug     bool
	conn      net.Conn
	state     int
	startTime time.Time
	errCount  int
}

func (this *Pop3Server) base64Encode(en string) string {
	src := []byte(en)
	maxLen := base64.StdEncoding.EncodedLen(len(src))
	dst := make([]byte, maxLen)
	base64.StdEncoding.Encode(dst, src)
	return string(dst)
}

func (this *Pop3Server) base64Decode(de string) string {
	dst, err := base64.StdEncoding.DecodeString(de)
	if err != nil {
		return ""
	}
	return string(dst)
}

func (this *Pop3Server) setState(state int) {
	this.state = state
}

func (this *Pop3Server) getState() int {
	return this.state
}

func (this *Pop3Server) D(a ...interface{}) (n int, err error) {
	return fmt.Println(a...)
}

func (this *Pop3Server) Debug(d bool) {
	this.debug = d
}

func (this *Pop3Server) write(code string) {

	info := fmt.Sprintf("%.3s %s%s", code, msgList[code], GO_EOL)
	_, err := this.conn.Write([]byte(info))

	if err != nil {
		log.Fatal(err)
	}
}

func (this *Pop3Server) getString() (string, error) {
	input, err := bufio.NewReader(this.conn).ReadString('\n')
	if err != nil {
		return "", err
	}
	inputTrim := strings.TrimSpace(input)
	return inputTrim, err
}

func (this *Pop3Server) getString0() (string, error) {
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

func (this *Pop3Server) close() {
	this.conn.Close()
}

func (this *Pop3Server) cmdCompare(input string, cmd int) bool {
	if strings.EqualFold(input, stateList[cmd]) {
		return true
	}
	return false
}

func (this *Pop3Server) stateCompare(input int, cmd int) bool {
	if input == cmd {
		return true
	}
	return false
}

func (this *Pop3Server) cmdHelo(input string) bool {
	inputN := strings.SplitN(input, " ", 2)

	if this.cmdCompare(inputN[0], CMD_HELO) {
		if len(inputN) < 2 {
			this.write(MSG_BAD_SYNTAX)
			return false
		}
		this.write(MSG_OK)
		return true
	}
	this.write(MSG_COMMAND_ERR)
	return false
}

func (this *Pop3Server) cmdEhlo(input string) bool {
	inputN := strings.SplitN(input, " ", 2)

	if this.cmdCompare(inputN[0], CMD_EHLO) {
		if len(inputN) < 2 {
			this.write(MSG_BAD_SYNTAX)
			return false
		}
		this.write(MSG_OK)
		return true
	}
	this.write(MSG_COMMAND_ERR)
	return false
}

func (this *Pop3Server) cmdQuit(input string) bool {
	if this.cmdCompare(input, CMD_QUIT) {
		this.write(MSG_BYE)
		this.close()
		return true
	}
	return false
}

func (this *Pop3Server) handle() {
	for {
		state := this.getState()
		input, _ := this.getString()

		fmt.Printf(input, state)
		break
	}
}

func (this *Pop3Server) start(conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(time.Minute * 180))
	defer conn.Close()
	this.conn = conn

	this.startTime = time.Now()

	this.write(MSG_INIT)
	this.setState(CMD_READY)

	this.handle()
}

func Start(port int) {
	pop3_port := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", pop3_port)
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

		srv := Pop3Server{}
		go srv.start(conn)
	}
}
