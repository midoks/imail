package pop3

import (
	"bufio"
	"fmt"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/libs"
	"log"
	"net"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	CMD_READY      = iota
	CMD_AUTH_PLAIN = iota
	CMD_USER       = iota
	CMD_PASS       = iota
	CMD_STAT       = iota
	CMD_LIST       = iota
	CMD_RETR       = iota
	CMD_DELE       = iota
	CMD_NOOP       = iota
	CMD_RSET       = iota
	CMD_TOP        = iota
	CMD_UIDL       = iota
	CMD_APOP       = iota
	CMD_QUIT       = iota
	CMD_CAPA       = iota
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
	MSG_RETR_DATA     = "%s octets\r\n%s\r\n"
	MSG_CAPA          = "Capability list follows"
	MSG_POS_DATA      = "%d %s"
	MSG_TOP_DATA      = "%s octets\r\n%s"
	MSG_AUTH_PLAIN    = "+\r\n"
)

var GO_EOL = GetGoEol()

func GetGoEol() string {
	if "windows" == runtime.GOOS {
		return "\r\n"
	}
	return "\n"
}

type Pop3Server struct {
	debug         bool
	conn          net.Conn
	state         int
	startTime     time.Time
	errCount      int
	recordCmdUser string
	recordCmdPass string

	// user id
	userID int64
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

func (this *Pop3Server) w(msg string) {
	_, err := this.conn.Write([]byte(msg))

	if err != nil {
		log.Fatal(err)
	}
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
			// fmt.Println(mailList)
			for i, m := range mailList {
				// fmt.Println(i, m.Size)
				this.writeInfo(MSG_STAT_OK, i+1, m.Size)
			}
			this.w(".\r\n")
			return true
		}
		// else if inputLen == 2 {
		// 	pos, err := strconv.ParseInt(inputN[1], 10, 64)
		// 	if err == nil {
		// 		if pos > 0 {
		// 			list, err := db.BoxPos(this.userID, pos)
		// 			if err == nil {
		// 				this.writeArgs(MSG_POS_DATA, pos, list[0]["mid"])
		// 				return true
		// 			}
		// 		}
		// 	}
		// }
		this.error(MSG_BAD_SYNTAX)
	}

	return false
}

func (this *Pop3Server) cmdUidl(input string) bool {
	inputN := strings.SplitN(input, " ", 2)
	fmt.Println("cmdUidl-:", inputN)
	if this.cmdCompare(inputN[0], CMD_UIDL) {
		inputLen := len(inputN)
		if inputLen == 2 {
			pos, err := strconv.ParseInt(inputN[1], 10, 64)
			if err == nil {

				if pos > 0 {
					list, err := db.MailListPosForPop(this.userID, pos)
					// fmt.Println(list)
					if err == nil {
						uid := strconv.FormatInt(list[0].Uid, 10)
						this.writeArgs(MSG_POS_DATA, pos, libs.Md5str(uid))
						return true
					}
				}
			}
		} else if inputLen == 1 {
			this.ok("")
			list, _ := db.MailListAllForPop(this.userID)
			for i := 1; i <= len(list); i++ {
				uid := strconv.FormatInt(list[i-1].Id, 10)
				t := fmt.Sprintf("%d %s\r\n", i, libs.Md5str(uid))
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
	// inputN := strings.SplitN(input, " ", 2)
	// if this.cmdCompare(inputN[0], CMD_TOP) {
	// 	if len(inputN) == 2 {
	// 		inputArgs := strings.SplitN(inputN[1], " ", 2)
	// 		if len(inputArgs) == 2 {
	// 			pos, err := strconv.ParseInt(inputArgs[0], 10, 64)
	// 			if err == nil {
	// 				line, err2 := strconv.ParseInt(inputArgs[1], 10, 64)
	// 				if err2 == nil {
	// 					content, size, err3 := db.BoxPosTop(this.userID, pos, line)
	// 					if err3 == nil {
	// 						// this.ok(content)
	// 						this.writeArgs(MSG_TOP_DATA, size, content)
	// 						return true
	// 					}
	// 				}
	// 			}
	// 		}
	// 	}
	// 	this.error(MSG_BAD_SYNTAX)
	// }
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
	fmt.Println("cmdAuthPlain", input)
	if this.cmdCompare(input, CMD_AUTH_PLAIN) {
		this.w(MSG_AUTH_PLAIN)
		return true
	}
	return false
}

func (this *Pop3Server) cmdParseAuthPlain(input string) bool {

	data, err := libs.Base64decode(input)
	if err == nil {
		this.D("pop3 src:", "cmdParseAuthPlain:", data)

		list := strings.SplitN(data, "\x00", 3)

		this.recordCmdUser = list[0]
		this.recordCmdPass = list[2]

		b := this.checkUserLogin()
		this.D("pop3 parse:", b, this.recordCmdUser, this.recordCmdPass)
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

		fmt.Println("pop3 cmd:", state, input)
		if this.cmdQuit(input) {
			break
		}

		if this.cmdCapa(input) {
		}

		if this.cmdAuthPlain(input) {
			this.setState(CMD_AUTH_PLAIN)
		}

		if this.stateCompare(state, CMD_AUTH_PLAIN) {
			if this.cmdParseAuthPlain(input) {
				this.setState(CMD_PASS)
			}
		}

		if this.stateCompare(state, CMD_READY) {
			if this.cmdUser(input) {
				this.setState(CMD_USER)
			}
		}

		if this.stateCompare(state, CMD_USER) {
			if this.cmdPass(input) {
				this.setState(CMD_PASS)
			}
		}

		if this.stateCompare(state, CMD_PASS) {

			if this.cmdStat(input) {
			}

			if this.cmdNoop(input) {
			}

			if this.cmdList(input) {
			}

			if this.cmdUidl(input) {
			}

			if this.cmdRetr(input) {
			}

			if this.cmdTop(input) {
			}
		}
	}
}

func (this *Pop3Server) start(conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(time.Minute * 10))
	defer conn.Close()
	this.conn = conn

	this.startTime = time.Now()

	this.ok(MSG_INIT)
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
		srv := Pop3Server{}
		go srv.start(conn)
	}
}
