package pop3

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/midoks/imail/internal/conf"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/log"
	"github.com/midoks/imail/internal/tools"
)

const (
	CMD_READY = iota
	CMD_AUTH_PLAIN
	CMD_USER
	CMD_PASS
	CMD_STAT
	CMD_LIST
	CMD_RETR
	CMD_DELE
	CMD_NOOP
	CMD_RSET
	CMD_TOP
	CMD_UIDL
	CMD_APOP
	CMD_QUIT
	CMD_CAPA
)

var stateList = map[int]string{
	CMD_READY:      "READY",
	CMD_USER:       "USER",
	CMD_PASS:       "PASS",
	CMD_STAT:       "STAT",
	CMD_LIST:       "LIST",
	CMD_RETR:       "RETR",
	CMD_DELE:       "DELE",
	CMD_NOOP:       "NOOP",
	CMD_RSET:       "RSET",
	CMD_TOP:        "TOP",
	CMD_UIDL:       "UIDL",
	CMD_APOP:       "APOP",
	CMD_QUIT:       "QUIT",
	CMD_CAPA:       "CAPA",
	CMD_AUTH_PLAIN: "AUTH PLAIN",
}

const (
	MSG_INIT          = "Welcome to coremail Mail Pop3 Server (imail)"
	MSG_OK            = "core mail"
	MSG_BAD_SYNTAX    = "500"
	MSG_LOGIN_OK      = "%d message(s) [%d byte(s)]"
	MSG_STAT_OK       = "%d %d"
	MSG_LOGIN_DISABLE = "Unable to log on"
	MSG_CMD_NOT_VALID = "Command not valid in this state"
	MSG_RETR_DATA     = "%s octets\r\n%s\r\n."
	MSG_CAPA          = "Capability list follows"
	MSG_POS_DATA      = "%d %s"
	MSG_TOP_DATA      = "%s octets\r\n%s"
	MSG_AUTH_PLAIN    = "+\r\n"
)

var GO_EOL = tools.GetGoEol()

type Pop3Server struct {
	debug         bool
	conn          net.Conn
	state         int
	startTime     time.Time
	errCount      int
	recordCmdUser string
	recordCmdPass string

	reader  *bufio.Reader
	writer  *bufio.Writer
	scanner *bufio.Scanner

	// user id
	userID int64

	LinkSSL   bool
	TLSConfig *tls.Config // Enable STARTTLS support.
}

func (this *Pop3Server) setState(state int) {
	this.state = state
}

func (this *Pop3Server) getState() int {
	return this.state
}

func (this *Pop3Server) D(args ...interface{}) {
	if this.LinkSSL {
		log.Debugf("[SSL]:%s", args...)
		return
	}

	if conf.Pop3.Debug {
		// fmt.Println(args...)
		log.Debug(args...)
	}
}

func (this *Pop3Server) Debug(d bool) {
	this.debug = d
}

func (this *Pop3Server) w(msg string) error {
	log := fmt.Sprintf("POP[w]:%s", msg)
	this.D(log)

	_, err := this.writer.Write([]byte(msg))
	this.writer.Flush()
	return err
}

func (this *Pop3Server) writeArgs(code string, args ...interface{}) {
	info := fmt.Sprintf("+OK "+code+"\r\n", args...)
	this.w(info)
}

func (this *Pop3Server) writeInfo(code string, args ...interface{}) {
	info := fmt.Sprintf(code+"\r\n", args...)
	this.w(info)
}

func (this *Pop3Server) ok(code string) {
	info := fmt.Sprintf("+OK %s\r\n", code)
	this.w(info)
}

func (this *Pop3Server) error(code string) {
	info := fmt.Sprintf("-ERR %s\r\n", code)
	this.w(info)
}

func (this *Pop3Server) getString() (string, error) {
	input, err := this.reader.ReadString('\n')

	if err != nil {
		return "", err
	}

	inputTrim := strings.TrimSpace(input)
	this.D("pop3[r]:", inputTrim, ":", err)
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

func (this *Pop3Server) checkUserLogin() bool {
	name := this.recordCmdUser
	pwd := strings.TrimSpace(this.recordCmdPass)

	isLogin, id := db.LoginWithCode(name, pwd)
	if !isLogin {
		return false
	}
	this.userID = id
	return true
}

func (this *Pop3Server) cmdUser(input string) bool {
	inputN := strings.SplitN(input, " ", 2)

	if this.cmdCompare(inputN[0], CMD_USER) {
		if len(inputN) < 2 {
			this.ok(MSG_BAD_SYNTAX)
			return false
		}

		this.recordCmdUser = strings.TrimSpace(inputN[1])
		this.ok(MSG_OK)
		return true
	}
	return false
}

func (this *Pop3Server) cmdPass(input string) bool {
	inputN := strings.SplitN(input, " ", 2)

	if this.cmdCompare(inputN[0], CMD_PASS) {
		if len(inputN) < 2 {
			this.ok(MSG_BAD_SYNTAX)
			return false
		}
		this.recordCmdPass = strings.TrimSpace(inputN[1])

		if this.checkUserLogin() {
			count, size := db.MailStatInfoForPop(this.userID)
			this.writeArgs(MSG_LOGIN_OK, count, size)
			return true
		}
		this.error(MSG_LOGIN_DISABLE)
		return false
	}
	return false
}

func (this *Pop3Server) cmdStat(input string) bool {
	if this.cmdCompare(input, CMD_STAT) {
		count, size := db.MailStatInfoForPop(this.userID)
		this.writeArgs(MSG_STAT_OK, count, size)
		return true
	}
	return false
}

func (this *Pop3Server) cmdList(input string) bool {
	inputN := strings.SplitN(input, " ", 2)

	if this.cmdCompare(inputN[0], CMD_LIST) {
		inputLen := len(inputN)
		if inputLen == 1 {
			count, size := db.MailStatInfoForPop(this.userID)
			this.writeArgs(MSG_STAT_OK, count, size)

			mailList := db.MailListForPop(this.userID)
			for i, m := range mailList {
				this.writeInfo(MSG_STAT_OK, i+1, m.Size)
			}
			this.w(".\r\n")
			return true
		}
		this.error(MSG_BAD_SYNTAX)
	}

	return false
}

func (this *Pop3Server) cmdUidl(input string) bool {
	inputN := strings.SplitN(input, " ", 2)
	if this.cmdCompare(inputN[0], CMD_UIDL) {
		inputLen := len(inputN)
		if inputLen == 2 {

			pos, err := strconv.ParseInt(inputN[1], 10, 64)
			if err == nil {
				if pos > 0 {
					list, err := db.MailListPosForPop(this.userID, pos)
					if err == nil && len(list) > 0 {
						for i := 1; i <= len(list); i++ {
							uid := strconv.FormatInt(list[i-1].Id, 10)
							this.writeArgs(MSG_POS_DATA, pos, tools.Md5(uid))
						}
						return true
					}

					this.w(".\r\n")
					return true
				}
			}
		} else if inputLen == 1 {

			this.ok("")
			list, _ := db.MailListAllForPop(this.userID)
			for i := 1; i <= len(list); i++ {
				uid := strconv.FormatInt(list[i-1].Id, 10)
				t := fmt.Sprintf("%d %s\r\n", i, tools.Md5(uid))
				this.w(t)
			}
			this.w(".\r\n")
			return true
		}
		this.error(MSG_BAD_SYNTAX)
	}
	return false
}

func (this *Pop3Server) cmdTop(input string) bool {
	return false
}

func (this *Pop3Server) cmdRetr(input string) bool {
	inputN := strings.SplitN(input, " ", 2)

	if this.cmdCompare(inputN[0], CMD_RETR) {
		if len(inputN) == 2 {
			pos, err := strconv.ParseInt(inputN[1], 10, 64)
			if err == nil {
				if pos > 0 {
					content, size, err := db.MailPosContentForPop(this.userID, pos)
					if err == nil {
						sizeStr := strconv.Itoa(size)
						this.writeArgs(MSG_RETR_DATA, sizeStr, content)
						return true
					}
				}
			}
		}
		this.error(MSG_BAD_SYNTAX)
	}
	return false
}

func (this *Pop3Server) cmdQuit(input string) bool {
	if this.cmdCompare(input, CMD_QUIT) {
		this.ok(MSG_OK)
		this.close()
		return true
	}
	return false
}

func (this *Pop3Server) cmdNoop(input string) bool {
	if this.cmdCompare(input, CMD_NOOP) {
		this.ok(MSG_OK)
		return true
	}
	return false
}

func (this *Pop3Server) cmdAuthPlain(input string) bool {

	if this.cmdCompare(input, CMD_AUTH_PLAIN) {
		this.D("pop3[cmdAuthPlain]:", input)
		this.w(MSG_AUTH_PLAIN)
		return true
	}
	return false
}

func (this *Pop3Server) cmdParseAuthPlain(input string) bool {

	data, err := tools.Base64decode(input)
	if err == nil {
		this.D("pop3[AuthPlain][Iuput]:", data)

		list := strings.SplitN(data, "\x00", 3)

		this.recordCmdUser = list[0]
		this.recordCmdPass = list[2]

		b := this.checkUserLogin()
		this.D("pop3[AuthPlain]:", b, this.recordCmdUser, this.recordCmdPass)
		if b {
			this.ok("Authentication successful")
			return true
		}
	}
	this.error(MSG_LOGIN_DISABLE)
	return false
}

func (this *Pop3Server) cmdCapa(input string) bool {
	if this.cmdCompare(input, CMD_CAPA) {
		this.ok(MSG_CAPA)
		this.w("TOP\r\n")
		this.w("USER\r\n")
		this.w("PIPELINING\r\n")
		this.w("UIDL\r\n")
		this.w("LANG\r\n")
		this.w("UTF8\r\n")
		this.w("SASL PLAIN\r\n")
		this.w("STLS\r\n")
		this.w(".\r\n")
		return true
	}
	return false
}

func (this *Pop3Server) handle() {
	for {
		state := this.getState()
		input, err := this.getString()

		if err != nil {
			break
		}

		this.D("pop3[cmd]:", state, input)
		if this.cmdQuit(input) {
			break
		}

		if this.cmdCapa(input) {
		} else if this.cmdAuthPlain(input) {
			this.setState(CMD_AUTH_PLAIN)
		} else if this.stateCompare(state, CMD_AUTH_PLAIN) {
			if this.cmdParseAuthPlain(input) {
				this.setState(CMD_PASS)
			}
		} else if this.stateCompare(state, CMD_READY) {
			if this.cmdUser(input) {
				this.setState(CMD_USER)
			}
		} else if this.stateCompare(state, CMD_USER) {
			if this.cmdPass(input) {
				this.setState(CMD_PASS)
			}
		} else if this.stateCompare(state, CMD_PASS) {

			if this.cmdStat(input) {
			} else if this.cmdNoop(input) {
			} else if this.cmdList(input) {
			} else if this.cmdUidl(input) {
			} else if this.cmdRetr(input) {
			} else if this.cmdTop(input) {
			}
		}
	}
}

func (this *Pop3Server) initTLSConfig() {
	this.TLSConfig = tools.InitAutoMakeTLSConfig()
}

func (this *Pop3Server) ready() {
	this.initTLSConfig()
}

func (this *Pop3Server) start(conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(time.Minute * 10))
	defer conn.Close()
	this.conn = conn

	this.reader = bufio.NewReader(conn)
	this.writer = bufio.NewWriter(conn)
	this.scanner = bufio.NewScanner(this.reader)

	this.startTime = time.Now()

	this.ok(MSG_INIT)
	this.setState(CMD_READY)

	this.handle()
}

func (this *Pop3Server) StartPort(port int) {
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		this.D("pop[start]:", err)
		return
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go this.start(conn)
	}
}

func (this *Pop3Server) StartSSLPort(port int) {
	this.ready()
	this.LinkSSL = true
	addr := fmt.Sprintf(":%d", port)
	ln, err := tls.Listen("tcp", addr, this.TLSConfig)
	if err != nil {
		this.D("pop[start][ssl]:", err)
		return
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go this.start(conn)
	}
}

func Start(port int) {
	srv := Pop3Server{}
	srv.StartPort(port)
}

func StartSSL(port int) {
	srv := Pop3Server{}
	srv.StartSSLPort(port)
}
