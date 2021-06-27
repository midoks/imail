package imap

import (
	"bufio"
	"fmt"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/libs"
	"io"
	"log"
	"net"
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
	CMD_NOOP       = iota
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
	CMD_NOOP:       "NOOP",
}

const (
	MSG_INIT           = "* OK [CAPABILITY IMAP4 IMAP4rev1 ID AUTH=PLAIN AUTH=LOGIN AUTH=XOAUTH2 NAMESPACE] imail ready"
	MSG_BAD_SYNTAX     = "%s BAD command not support"
	MSG_LOGIN_OK       = "%s OK LOGIN completed"
	MSG_LOGOUT_OK      = "%s OK LOGOUT completed"
	MSG_LOGIN_DISABLE  = "%s NO LOGIN Login error password error"
	MSG_CMD_NOT_VALID  = "Command not valid in this state"
	MSG_LOGOUT         = "* BYE IMAP4rev1 Server logging out"
	MSG_COMPLELED      = "%s OK %s completed"
	MSG_COMPLELED_LIST = "* %s %s %s"
)

var GO_EOL = libs.GetGoEol()

// An IMAP reader.
// type Reader struct {
// 	MaxLiteralSize uint32 // The maximum literal size.

// 	reader

// 	continues chan<- bool

// 	brackets   int
// 	inRespCode bool
// }

type ImapServer struct {
	io.Reader
	io.RuneScanner
	// *Writer
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

	selectBox string
	// commands  map[int]HandlerFactory
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

func (this *ImapServer) w(msg string) error {
	log := fmt.Sprintf("imap[w]:%s", msg)
	this.D(log)

	_, err := this.writer.Write([]byte(msg))
	this.writer.Flush()
	return err
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

func (this *ImapServer) getString(state int) (string, error) {
	// if state == CMD_DATA {
	// 	return "", nil
	// }

	fmt.Println(state)

	input, err := this.reader.ReadString('\n')
	inputTrim := strings.TrimSpace(input)
	this.D("imap[r]:", inputTrim, ":", err)
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

		user := strings.Trim(inputN[2], "\"")
		pwd := strings.Trim(inputN[3], "\"")

		fmt.Println(user, pwd)

		isLogin, id := db.LoginWithCode(user, pwd)
		if isLogin {
			this.userID = id
			this.writeArgs(MSG_LOGIN_OK, inputN[0])
			return true
		}
		this.writeArgs(MSG_LOGIN_DISABLE, inputN[0])
	}
	return false
}

func (this *ImapServer) cmdCapabitity(input string) bool {
	inputN := strings.SplitN(input, " ", 2)

	fmt.Println(inputN)
	if len(inputN) == 2 {
		if this.cmdCompare(inputN[1], CMD_CAPABILITY) {
			this.writeArgs("* OK Coremail System IMap Server Ready(imail)")
			this.w("* CAPABILITY IMAP4rev1 XLIST SPECIAL-USE ID LITERAL+ STARTTLS XAPPLEPUSHSERVICE UIDPLUS X-CM-EXT-1\r\n")
			this.writeArgs(MSG_COMPLELED, inputN[0], inputN[1])
			return true
		}
	}
	return false
}

func (this *ImapServer) cmdId(input string) bool {
	inputN := strings.SplitN(input, " ", 3)
	if len(inputN) == 3 {
		if this.cmdCompare(inputN[1], CMD_ID) {
			this.writeArgs(MSG_COMPLELED, inputN[0], inputN[1])
			return true
		}
	}
	return false
}

func (this *ImapServer) cmdList(input string) bool {
	inputN := strings.SplitN(input, " ", 4)
	if len(inputN) == 4 {
		if this.cmdCompare(inputN[1], CMD_LIST) {
			this.writeArgs("* LIST (\\NoSelect \\HasChildren) \"/\" \"&UXZO1mWHTvZZOQ-\"")
			this.writeArgs("* LIST (\\HasChildren) \"/\" \"INBOX\"")
			this.writeArgs("* LIST (\\HasChildren) \"/\" \"Sent Messages\"")
			this.writeArgs("* LIST (\\HasChildren) \"/\" \"Drafts\"")
			this.writeArgs("* LIST (\\HasChildren) \"/\" \"Deleted Messages\"")
			this.writeArgs("* LIST (\\HasChildren) \"/\" \"Junk\"")
			this.writeArgs(MSG_COMPLELED, inputN[0], inputN[1])
			return true
		}
	}
	return false
}

func (this *ImapServer) cmdStatus(input string) bool {
	inputN := strings.SplitN(input, " ", 4)
	if len(inputN) == 4 {
		if this.cmdCompare(inputN[1], CMD_STATUS) {
			this.writeArgs(MSG_COMPLELED_LIST, inputN[0], inputN[1], inputN[3])
			this.writeArgs(MSG_COMPLELED, inputN[0], inputN[1])
			return true
		}
	}
	return false
}

func (this *ImapServer) cmdSelect(input string) bool {
	inputN := strings.SplitN(input, " ", 3)
	if len(inputN) == 3 {
		if this.cmdCompare(inputN[1], CMD_SELECT) {
			this.selectBox = strings.Trim(inputN[2], "\"")
			msgCount, _ := db.BoxUserMessageCountByClassName(this.userID, this.selectBox)
			this.writeArgs("* %d EXISTS", msgCount)
			this.writeArgs("* 0 RECENT")
			this.writeArgs("* OK [UIDVALIDITY 1] UIDs valid")
			this.writeArgs("* FLAGS (\\Answered \\Seen \\Deleted \\Draft \\Flagged)")
			this.writeArgs("* OK [PERMANENTFLAGS (\\Answered \\Seen \\Deleted \\Draft \\Flagged)] Limited")
			this.writeArgs("%s OK [READ-WRITE] %s completed", inputN[0], inputN[1])
			return true
		}
	}
	return false
}

func (this *ImapServer) cmdFecth(input string) bool {
	inputN := strings.SplitN(input, " ", 4)
	if len(inputN) == 4 {
		if this.cmdCompare(inputN[1], CMD_FETCH) {
			fmt.Println("cmdFecth", inputN)

			mailList := db.MailListForPop(this.userID)
			for i, m := range mailList {
				this.writeArgs("* %d FETCH (UID %d)", i+1, m.Id)
			}

			this.writeArgs(MSG_COMPLELED, inputN[0], inputN[1])
			return true
		}
	}
	return false
}

func (this *ImapServer) cmdUid(input string) bool {
	inputN := strings.SplitN(input, " ", 5)
	if len(inputN) == 5 {
		if this.cmdCompare(inputN[1], CMD_UID) {
			fmt.Println("cmdUid", inputN)

			this.writeArgs(MSG_COMPLELED, inputN[0], inputN[1])
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
		state := this.state
		input, err := this.getString(state)
		fmt.Println("input:", input, "err", err)

		if err != nil {
			this.close()
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

			if this.cmdFecth(input) {

			}

			if this.cmdLogout(input) {
				break
			}
		}

	}
}

func (this *ImapServer) start(conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(time.Minute * 10))
	defer conn.Close()
	this.conn = conn

	this.reader = bufio.NewReader(conn)
	this.writer = bufio.NewWriter(conn)
	this.scanner = bufio.NewScanner(this.reader)

	this.startTime = time.Now()

	this.ok(MSG_INIT)
	this.setState(CMD_READY)

	// this.commands = map[int]HandlerFactory{
	// 	CMD_FETCH:  func() Handler { return &cmd.Fetch{} },
	// 	CMD_NOOP:   func() Handler { return &cmd.Noop{} },
	// 	CMD_UID:    func() Handler { return &cmd.Uid{} },
	// 	CMD_LIST:   func() Handler { return &cmd.List{} },
	// 	CMD_STATUS: func() Handler { return &cmd.Status{} },
	// 	CMD_SELECT: func() Handler { return &cmd.Select{} },
	// }
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
