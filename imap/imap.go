package imap

import (
	"bufio"
	"fmt"
	"github.com/midoks/imail/app/models"
	// "github.com/midoks/imail/libs"
	"log"
	"net"
	"runtime"
	// "strconv"
	"strings"
	"time"
)

const (
	CMD_READY      = iota
	CMD_AUTH       = iota
	CMD_LIST       = iota
	CMD_LOGOUT     = iota
	CMD_CAPABILITY = iota
	CMD_ID         = iota
	CMD_STATUS     = iota
	CMD_SELECT     = iota
	CMD_FETCH      = iota
	CMD_UID        = iota
)

var stateList = map[int]string{
	CMD_READY:      "READY",
	CMD_AUTH:       "LOGIN",
	CMD_LOGOUT:     "LOGOUT",
	CMD_LIST:       "LIST",
	CMD_CAPABILITY: "CAPABILITY",
	CMD_ID:         "ID",
	CMD_STATUS:     "STATUS",
	CMD_SELECT:     "SELECT",
	CMD_FETCH:      "FETCH",
	CMD_UID:        "UID",
}

const (
	MSG_INIT          = "* OK Coremail System IMap Server Ready(imail)"
	MSG_BAD_SYNTAX    = "%s BAD command not support"
	MSG_LOGIN_OK      = "%s OK LOGIN completed"
	MSG_LOGOUT_OK     = "%s OK LOGOUT completed"
	MSG_LOGIN_DISABLE = "%s NO LOGIN Login error password error"
	MSG_CMD_NOT_VALID = "Command not valid in this state"
	MSG_LOGOUT        = "* BYE IMAP4rev1 Server logging out"
)

var GO_EOL = GetGoEol()

func GetGoEol() string {
	if "windows" == runtime.GOOS {
		return "\r\n"
	}
	return "\n"
}

type ImapServer struct {
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

func (this *ImapServer) setState(state int) {
	this.state = state
}

func (this *ImapServer) getState() int {
	return this.state
}

func (this *ImapServer) D(a ...interface{}) (n int, err error) {
	return fmt.Println(a...)
}

func (this *ImapServer) Debug(d bool) {
	this.debug = d
}

func (this *ImapServer) w(msg string) {
	// fmt.Println("w[debug]:", msg)
	_, err := this.conn.Write([]byte(msg))

	if err != nil {
		log.Fatal(err)
	}
}

func (this *ImapServer) writeArgs(code string, args ...interface{}) {
	info := fmt.Sprintf(code+"\r\n", args...)
	this.w(info)
}

func (this *ImapServer) ok(code string) {
	info := fmt.Sprintf("%s\r\n", code)
	this.w(info)
}

func (this *ImapServer) error(code string) {
	info := fmt.Sprintf("%s\r\n", code)
	this.w(info)
}

func (this *ImapServer) getString() (string, error) {
	input, err := bufio.NewReader(this.conn).ReadString('\n')
	if err != nil {
		return "", err
	}
	inputTrim := strings.TrimSpace(input)
	return inputTrim, err
}

func (this *ImapServer) getString0() (string, error) {
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

func (this *ImapServer) close() {
	this.conn.Close()
}

func (this *ImapServer) cmdCompare(input string, cmd int) bool {
	if strings.EqualFold(input, stateList[cmd]) {
		return true
	}
	return false
}

func (this *ImapServer) stateCompare(input int, cmd int) bool {
	if input == cmd {
		return true
	}
	return false
}

func (this *ImapServer) cmdAuth(input string) bool {
	inputN := strings.SplitN(input, " ", 4)
	if this.cmdCompare(inputN[1], CMD_AUTH) {
		if len(inputN) < 4 {
			this.writeArgs(MSG_BAD_SYNTAX, inputN[0])
			return false
		}

		user := inputN[2]
		pwd := strings.Trim(inputN[3], "\"")

		isLogin, id := models.UserLogin(user, pwd)
		if isLogin {
			this.userID = id
			this.writeArgs(MSG_LOGIN_OK, inputN[0])
			return true
		}
		this.writeArgs(MSG_LOGIN_DISABLE, inputN[0])
	}
	return false
}

func (this *ImapServer) cmdStatus(input string) bool {
	inputN := strings.SplitN(input, " ", 4)

	if len(inputN) == 4 {
		if this.cmdCompare(inputN[1], CMD_STATUS) {
			fmt.Println("cmd_list:", inputN)
			this.writeArgs("* %s %s (MESSAGES 122 RECENT 0 UNSEEN 0)", inputN[1], inputN[2])
			this.writeArgs("%s OK %s completed", inputN[0], inputN[1])
			return true
		}
	}
	return false
}

func (this *ImapServer) cmdSelect(input string) bool {
	inputN := strings.SplitN(input, " ", 3)

	if len(inputN) == 3 {
		if this.cmdCompare(inputN[1], CMD_SELECT) {
			fmt.Println("cmd_list:", inputN)
			this.w("* 122 EXISTS\r\n")
			this.w("* 1 RECENT\r\n")
			this.w("* OK [UIDVALIDITY 1] UIDs valid\r\n")
			this.w("* FLAGS (\\Answered\\Seen \\Deleted \\Draft \\Flagged)\r\n")
			this.w("* OK [PERMANENTFLAGS (\\Answered \\Seen \\Deleted \\Draft \\Flagged)] Limited\r\n")
			this.writeArgs("%s OK [READ-WRITE] %s completed", inputN[0], inputN[1])
			return true
		}
	}
	return false
}

func (this *ImapServer) cmdFetch(input string) bool {
	inputN := strings.SplitN(input, " ", 4)

	if len(inputN) == 4 {
		if this.cmdCompare(inputN[1], CMD_FETCH) {
			this.w("* 1 FETCH (UID 1320476750)\r\n")
			this.w("* 2 FETCH (UID 1320476751)\r\n")
			this.writeArgs("%s OK %s completed", inputN[0], inputN[1])
			return true
		}
	}
	return false
}

func (this *ImapServer) cmdUid(input string) bool {
	inputN := strings.SplitN(input, " ", 5)

	if len(inputN) == 4 {
		if this.cmdCompare(inputN[1], CMD_UID) {
			this.w("* 1 FETCH (UID 1320476750)\r\n")
			this.w("* 2 FETCH (UID 1320476751)\r\n")
			this.writeArgs("%s OK %s completed", inputN[0], inputN[1])
			return true
		}
	}
	return false
}

func (this *ImapServer) cmdList(input string) bool {
	inputN := strings.SplitN(input, " ", 4)
	if len(inputN) == 4 {
		if this.cmdCompare(inputN[1], CMD_LIST) {
			list, err := models.ClassGetByUid(this.userID)
			if err == nil {
				for i := 1; i <= len(list); i++ {
					this.writeArgs("* LIST (\\%s) \"/\" \"%s\"", list[i-1]["tag"], list[i-1]["name"])
				}
				this.writeArgs("%s OK %s completed", inputN[0], inputN[1])
				return true
			}
		}
	}
	return false
}

func (this *ImapServer) cmdCapabitity(input string) bool {
	inputN := strings.SplitN(input, " ", 2)

	if len(inputN) == 2 {
		if this.cmdCompare(inputN[1], CMD_CAPABILITY) {
			this.w("* OK Coremail System IMap Server Ready(imail)\r\n")
			this.w("* CAPABILITY IMAP4rev1 XLIST SPECIAL-USE ID LITERAL+ STARTTLS XAPPLEPUSHSERVICE UIDPLUS X-CM-EXT-1\r\n")
			this.writeArgs("%s OK CAPABILITY completed\r\n", inputN[0])
			return true
		}

	}
	return false
}

func (this *ImapServer) cmdId(input string) bool {
	inputN := strings.SplitN(input, " ", 3)
	if len(inputN) == 3 {
		if this.cmdCompare(inputN[1], CMD_ID) {
			this.writeArgs("%s OK %s completed\r\n", inputN[0], inputN[1])
			return true
		}
	}
	return false
}

func (this *ImapServer) cmdLogout(input string) bool {
	inputN := strings.SplitN(input, " ", 2)

	if len(inputN) == 2 {
		if this.cmdCompare(inputN[1], CMD_LOGOUT) {
			this.writeArgs(MSG_LOGOUT)
			this.writeArgs(MSG_LOGOUT_OK, inputN[0])
			this.close()
			return true
		}
	}
	return false
}

func (this *ImapServer) handle() {
	for {
		state := this.getState()
		input, err := this.getString()
		if err != nil {
			break
		}

		fmt.Println("imap:", state, input)

		if this.cmdLogout(input) {
			break
		}

		if this.cmdCapabitity(input) {
		}

		if this.cmdId(input) {
		}

		if this.cmdAuth(input) {
			this.setState(CMD_AUTH)
		}

		if this.stateCompare(state, CMD_AUTH) {

			if this.cmdList(input) {
			}

			if this.cmdStatus(input) {
			}

			if this.cmdSelect(input) {
			}

			if this.cmdFetch(input) {
			}

			if this.cmdUid(input) {
			}

		}
	}
}

func (this *ImapServer) start(conn net.Conn) {
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
		srv := ImapServer{}
		go srv.start(conn)
	}
}
