package smtpd

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/midoks/imail/app/models"
	"github.com/midoks/imail/libs"
	"log"
	"net"
	"runtime"
	"strings"
	"time"
)

// There are two ways
// CMD_METHO_SEND is the data delivered by other SMTP servers.
// CMD_METHO_USER is data sent by users.
const (
	CMD_METHO_SEND = iota
	CMD_METHO_USER = iota
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
	MSG_OK              = "250.ok"
	MSG_MAIL_OK         = "250"
	MSG_BYE             = "221"
	MSG_BAD_SYNTAX      = "500"
	MSG_COMMAND_ERR     = "502"
	MSG_BAD_USER        = "505"
	MSG_BAD_OPEN_RELAY  = "505.open.relay"
	MSG_BAD_MAIL_ADDR   = "554"
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
	MSG_AUTH_LOGIN_USER: "dXNlcm5hbWU6",
	MSG_AUTH_LOGIN_PWD:  "UGFzc3dvcmQ6",
	MSG_AUTH_OK:         "Authentication successful",
	MSG_AUTH_FAIL:       "Error: authentication failed",
	MSG_MAIL_OK:         "Mail OK",
	MSG_DATA:            "End data with <CR><LF>.<CR><LF>",
	MSG_BAD_SYNTAX:      "Error: bad syntax",
	MSG_BAD_USER:        "Invalid User",
	MSG_BAD_OPEN_RELAY:  "Anonymous forwarding is not supported",
	MSG_BAD_MAIL_ADDR:   "The sender of the envelope does not match the sender of the letter.",
}

var GO_EOL = GetGoEol()

func GetGoEol() string {
	if "windows" == runtime.GOOS {
		return "\r\n"
	}
	return "\n"
}

type SmtpdServer struct {
	method            int
	debug             bool
	conn              net.Conn
	state             int
	startTime         time.Time
	errCount          int
	loginUser         string
	loginPwd          string
	recordCmdHelo     string
	recordCmdMailFrom string
	recordcmdRcptTo   string
	recordCmdData     string

	//DB DATA
	userID int64
}

func (this *SmtpdServer) base64Encode(en string) string {
	src := []byte(en)
	maxLen := base64.StdEncoding.EncodedLen(len(src))
	dst := make([]byte, maxLen)
	base64.StdEncoding.Encode(dst, src)
	return string(dst)
}

func (this *SmtpdServer) base64Decode(de string) string {
	dst, err := base64.StdEncoding.DecodeString(de)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(dst)
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

func (this *SmtpdServer) w(msg string) {
	_, err := this.conn.Write([]byte(msg))

	if err != nil {
		log.Fatal(err)
	}
}

func (this *SmtpdServer) write(code string) {
	info := fmt.Sprintf("%.3s %s%s", code, msgList[code], GO_EOL)
	_, err := this.conn.Write([]byte(info))

	if err != nil {
		log.Fatal(err)
	}
}

func (this *SmtpdServer) getString(state int) (string, error) {
	if state == CMD_DATA {
		return "", nil
	}

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
	this.method = CMD_METHO_USER
	if len(inputN) == 2 {
		if this.cmdCompare(inputN[0], CMD_HELO) {
			this.write(MSG_OK)
			return true
		}
	}
	return false
}

func (this *SmtpdServer) cmdEhlo(input string) bool {
	this.method = CMD_METHO_SEND
	inputN := strings.SplitN(input, " ", 2)
	if len(inputN) == 2 {

		if this.cmdCompare(inputN[0], CMD_EHLO) {
			// this.write(MSG_OK)
			this.w("250-mail\r\n")
			this.w("250-PIPELINING\r\n")
			this.w("250-AUTH LOGIN PLAIN\r\n")
			this.w("250-AUTH=LOGIN PLAIN\r\n")
			this.w("250-coremail 1Uxr2xKj7kG0xkI17xGrU7I0s8FY2U3Uj8Cz28x1UUUUU7Ic2I0Y2UFRbmXhUCa0xDrUUUUj\r\n")
			this.w("250-STARTTLS\r\n")
			this.w("250-SIZE 73400320\r\n")
			this.w("250 8BITMIME\r\n")
			return true
		}
	}
	return false
}

func (this *SmtpdServer) cmdAuthLogin(input string) bool {
	if this.cmdCompare(input, CMD_AUTH_LOGIN) {
		this.write(MSG_AUTH_LOGIN_USER)
		return true
	}
	this.write(MSG_BAD_SYNTAX)
	return false
}

func (this *SmtpdServer) checkUserLogin() bool {
	name := this.loginUser
	pwd := this.loginPwd

	info, err := models.UserGetByName(name)

	if err != nil {
		return false
	}

	pwdStr := libs.Md5str(pwd)
	if pwdStr != info.Password {
		return false
	}

	return true
}

func (this *SmtpdServer) cmdAuthLoginUser(input string) bool {

	user := this.base64Decode(input)
	this.loginUser = user
	this.D(this.loginUser)
	this.write(MSG_AUTH_LOGIN_PWD)
	return true
}

func (this *SmtpdServer) cmdAuthLoginPwd(input string) bool {
	pwd := this.base64Decode(input)
	this.loginPwd = pwd

	// fmt.Println(this.loginPwd)
	if this.checkUserLogin() {
		this.write(MSG_AUTH_OK)
		return true
	}
	this.write(MSG_AUTH_FAIL)
	return false
}

func (this *SmtpdServer) cmdMailFrom(input string) bool {
	inputN := strings.SplitN(input, ":", 2)

	if len(inputN) == 2 {
		if this.cmdCompare(inputN[0], CMD_MAIL_FROM) {

			inputN[1] = strings.TrimSpace(inputN[1])

			if !libs.CheckStandardMail(inputN[1]) {
				this.write(MSG_BAD_SYNTAX)
				return false
			}

			mailFrom := libs.GetRealMail(inputN[1])
			if !libs.IsEmailRe(mailFrom) {
				this.write(MSG_BAD_USER)
				return false
			}

			if this.method == CMD_METHO_USER {
				info := strings.Split(mailFrom, "@")
				mdomain := beego.AppConfig.String("mail.domain")
				if !strings.EqualFold(mdomain, info[1]) {
					this.write(MSG_BAD_MAIL_ADDR)
					return false
				}

				user, err := models.UserGetByName(info[0])
				if err != nil {
					this.write(MSG_BAD_USER)
					return false
				}
				this.userID = user.Id
			}

			this.recordCmdMailFrom = mailFrom
			this.write(MSG_MAIL_OK)
			return true
		}
	}
	this.write(MSG_BAD_SYNTAX)
	return false
}

func (this *SmtpdServer) cmdRcptTo(input string) bool {
	inputN := strings.SplitN(input, ":", 2)
	if len(inputN) == 2 {
		if this.cmdCompare(inputN[0], CMD_RCPT_TO) {
			inputN[1] = strings.TrimSpace(inputN[1])

			if !libs.CheckStandardMail(inputN[1]) {
				this.write(MSG_BAD_SYNTAX)
				return false
			}

			rcptTo := libs.GetRealMail(inputN[1])

			if !libs.IsEmailRe(rcptTo) {
				this.write(MSG_BAD_USER)
				return false
			}
			this.recordcmdRcptTo = rcptTo

			if this.method == CMD_METHO_SEND {
				info := strings.Split(rcptTo, "@")
				mdomain := beego.AppConfig.String("mail.domain")
				if !strings.EqualFold(mdomain, info[1]) {
					this.write(MSG_BAD_OPEN_RELAY)
					return false
				}

				user, err := models.UserGetByName(info[0])
				if err != nil {
					this.write(MSG_BAD_USER)
					return false
				}
				this.userID = user.Id
			}

			this.write(MSG_MAIL_OK)
			return true
		}
	}
	this.write(MSG_BAD_SYNTAX)
	return false
}

func (this *SmtpdServer) cmdData(input string) bool {
	if this.cmdCompare(input, CMD_DATA) {
		this.write(MSG_DATA)
		return true
	}
	this.write(MSG_BAD_SYNTAX)
	return false
}

func (this *SmtpdServer) cmdDataAccept() bool {
	var content string
	content = ""
	for {

		b := make([]byte, 4096)
		n, _ := this.conn.Read(b[0:])

		line := strings.TrimSpace(string(b[:n]))
		content += fmt.Sprintf("%s", line)

		last := line[len(line)-1:]

		if strings.EqualFold(last, ".") {
			content = strings.TrimSpace(content[0 : len(content)-1])
			this.write(MSG_DATA)
			break
		}
	}

	if this.method == CMD_METHO_SEND {
		mid, err := models.MailPush(this.recordCmdMailFrom, this.recordcmdRcptTo, content)
		if err != nil {
			return false
		}
		models.BoxAdd(this.userID, mid, 0, len(content))
	}

	if this.method == CMD_METHO_USER {
		mid, err := models.MailPush(this.recordCmdMailFrom, this.recordcmdRcptTo, content)
		if err != nil {
			return false
		}
		models.BoxAdd(this.userID, mid, 1, len(content))
	}

	return true
}

func (this *SmtpdServer) cmdQuit(input string) bool {
	if this.cmdCompare(input, CMD_QUIT) {
		this.write(MSG_BYE)
		this.close()
		return true
	}
	return false
}

func (this *SmtpdServer) handle() {
	for {
		state := this.getState()

		input, err := this.getString(state)
		if err != nil {
			break
		}

		this.D(input, state, stateList[state])

		//CMD_READY
		if this.stateCompare(state, CMD_READY) {
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
		}

		//CMD_HELO
		if this.stateCompare(state, CMD_HELO) {
			if this.cmdQuit(input) {
				break
			}

			if this.cmdAuthLogin(input) {
				this.setState(CMD_AUTH_LOGIN)
			}
		}

		//CMD_EHLO
		if this.stateCompare(state, CMD_EHLO) {
			if this.cmdQuit(input) {
				break
			}

			if this.cmdMailFrom(input) {
				this.setState(CMD_MAIL_FROM)
			}
		}

		//CMD_AUTH_LOGIN
		if this.stateCompare(state, CMD_AUTH_LOGIN) || this.stateCompare(state, CMD_MAIL_FROM) {
			if this.cmdAuthLoginUser(input) {
				this.setState(CMD_AUTH_LOGIN_USER)
			}
		}

		//CMD_AUTH_LOGIN_USER
		if this.stateCompare(state, CMD_AUTH_LOGIN_USER) {
			if this.cmdQuit(input) {
				break
			}
			if this.cmdAuthLoginPwd(input) {
				this.setState(CMD_AUTH_LOGIN_PWD)
			}
		}

		//CMD_AUTH_LOGIN_PWD
		if this.stateCompare(state, CMD_AUTH_LOGIN_PWD) {
			if this.cmdQuit(input) {
				break
			}

			if this.cmdMailFrom(input) {
				this.setState(CMD_MAIL_FROM)
			}
		}

		//CMD_MAIL_FROM
		if this.stateCompare(state, CMD_MAIL_FROM) {
			if this.cmdQuit(input) {
				break
			}

			if this.cmdRcptTo(input) {
				this.setState(CMD_RCPT_TO)
			}
		}

		//CMD_RCPT_TO
		if this.stateCompare(state, CMD_RCPT_TO) {
			if this.cmdQuit(input) {
				break
			}

			if this.cmdData(input) {
				this.setState(CMD_DATA)
			}
		}

		//CMD_DATA
		if this.stateCompare(state, CMD_DATA) {
			if this.cmdDataAccept() {
				this.setState(CMD_DATA_END)
			}
		}

		//CMD_DATA_END
		if this.stateCompare(state, CMD_DATA_END) {
			if this.cmdQuit(input) {
				break
			}
		}
	}
}

func (this *SmtpdServer) start(conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(time.Minute * 60))
	this.conn = conn
	defer conn.Close()
	this.startTime = time.Now()

	this.write(MSG_INIT)
	this.setState(CMD_READY)

	this.handle()
}

func Start(port int) {
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
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

		srv := SmtpdServer{}
		go srv.start(conn)
	}
}
