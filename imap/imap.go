package imap

import (
	"bufio"
	"fmt"
	"github.com/midoks/imail/app/models"
	"github.com/midoks/imail/libs/utf7"
	// "github.com/midoks/imail/libs"
	"log"
	"net"
	"runtime"
	"strconv"
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
	MSG_INIT          = "* OK Coremail System IMap Server Ready(imail)"
	MSG_BAD_SYNTAX    = "%s BAD command not support"
	MSG_LOGIN_OK      = "%s OK LOGIN completed"
	MSG_LOGOUT_OK     = "%s OK LOGOUT completed"
	MSG_LOGIN_DISABLE = "%s NO LOGIN Login error password error"
	MSG_CMD_NOT_VALID = "Command not valid in this state"
	MSG_LOGOUT        = "* BYE IMAP4rev1 Server logging out"
	MSG_COMPLELED     = "%s OK %s completed"
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

	selectBox string

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

func (this *ImapServer) parseArgs(input string) string {
	input = strings.TrimSpace(input)
	input = strings.Trim(input, "()")

	inputN := strings.Split(input, " ")
	list := make(map[string]int64)

	for i := 0; i < len(inputN); i++ {
		if strings.EqualFold(inputN[i], "messages") {
			count, _ := models.BoxUserMessageCountByClassName(this.userID, this.selectBox)
			list[inputN[i]] = count
		}
		if strings.EqualFold(inputN[i], "recent") {
			list[inputN[i]] = 0
		}

		if strings.EqualFold(inputN[i], "unseen") {
			list[inputN[i]] = 0
		}

	}

	out := ""
	for i := 0; i < len(inputN); i++ {
		// fmt.Println(i, inputN[i], list[inputN[i]])
		out += fmt.Sprintf("%s %d ", inputN[i], list[inputN[i]])
	}

	out = fmt.Sprintf("( %s )", out)
	return out
}

func (this *ImapServer) parseArgsConent(format string, mid string) string {
	format = strings.TrimSpace(format)
	format = strings.Trim(format, "()")

	inputN := strings.Split(format, " ")
	list := make(map[string]interface{})

	midInt64, _ := strconv.ParseInt(mid, 10, 64)
	s, _ := models.MailById(midInt64)
	content := s["content"].(string)

	bufferedBody := bufio.NewReader(strings.NewReader(content))
	header, err := ReadHeader(bufferedBody)

	// fmt.Println("headerString:", headerString)
	if err != nil {
		fmt.Errorf("Expected no error while reading mail, got:", err)
	}
	bs, err := FetchBodyStructure(header, bufferedBody, true)
	if err == nil {
		fmt.Println("FetchBodyStructure------333:", bs)
		bs.ToString()
	} else {
		fmt.Println(err)
	}

	// contentN := strings.Split(content, "\n\n")
	contentL := strings.Split(content, "\n")
	// fmt.Println("cmdCompare::--------\r\n", contentN, len(contentN))
	// fmt.Println("cmdCompare::--------\r\n")

	for i := 0; i < len(inputN); i++ {

		if strings.EqualFold(inputN[i], "uid") {
			// list[inputN[i]] = libs.Md5str(mid)
			list[inputN[i]] = mid
		}

		if strings.EqualFold(inputN[i], "flags") {
			list[inputN[i]] = "(\\Seen)"
		}

		if strings.EqualFold(inputN[i], "rfc822.size") {
			nnn := fmt.Sprintf("%d", len(content))
			list[inputN[i]] = nnn
		}

		if strings.EqualFold(inputN[i], "bodystructure") {
			cccc := fmt.Sprintf("(\"text\" \"plain\" (\"charset\" \"utf-8\") NIL NIL \"8bit\" %d %d NIL NIL NIL)", len(content), len(contentL[0]))
			// cccc := "(  )"

			// ccc2 := "((\"text\" \"plain\" (\"charset\" \"UTF-8\") NIL NIL \"quoted-printable\" 4542 104 NIL NIL NIL)(\"text\" \"html\" (\"charset\" \"UTF-8\") NIL NIL \"quoted-printable\" 43308 574 NIL NIL NIL) \"alternative\" (\"boundary\" \"--==_mimepart_5d09e7387efec_127483fd2fc2449c43048322e7\" \"charset\" \"UTF-8\") NIL NIL)"
			list[inputN[i]] = cccc

			list[inputN[i]] = bs.ToString()
		}

		if strings.EqualFold(inputN[i], "body.peek[header]") {
			headerString, _ := ReadHeaderString(bufio.NewReader(strings.NewReader(content)))
			list["body[header]"] = fmt.Sprintf("{%d}\r\n%s\r\n", len(headerString), headerString)
			// list[inputN[i]] = "{1218} \r\nTo: \"midoks@163.com\" <midoks@163.com> \r\nFrom:  <report-noreply@jiankongbao.com>\r\nSubject: 123123\r\nMessage-ID: <80d0b8ee122340ceb665ad1bf5220a42@localhost.localdomain>"
		}

		if strings.EqualFold(inputN[i], "body.peek[]") {
			list[inputN[i]] = fmt.Sprintf("{%d}\r\n%s\r\n", len(content), content)
			// list[inputN[i]] = "{1218} \r\nTo: \"midoks@163.com\" <midoks@163.com> \r\nFrom:  <report-noreply@jiankongbao.com>\r\nSubject: 123123\r\nMessage-ID: <80d0b8ee122340ceb665ad1bf5220a42@localhost.localdomain>"
		}
	}

	out := ""
	for i := 0; i < len(inputN); i++ {
		fmt.Println(i, inputN[i], list[inputN[i]])

		if strings.EqualFold(inputN[i], "body.peek[header]") {
			out += fmt.Sprintf("%s %s", "body[header]", list["body[header]"].(string))
		} else if strings.EqualFold(inputN[i], "body[]") {
			out += fmt.Sprintf("%s %s", "body[]", list["body[]"].(string))
		} else {
			out += fmt.Sprintf("%s %s", inputN[i], list[inputN[i]].(string))
		}
	}

	out = fmt.Sprintf("(%s)", out)
	fmt.Println(out)
	return out
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
			this.selectBox = strings.Trim(inputN[2], "\"")
			outArgs := this.parseArgs(inputN[3])
			this.writeArgs("* %s %s %s", inputN[1], inputN[2], outArgs)
			this.writeArgs(MSG_COMPLELED, inputN[0], inputN[1])
			return true
		}
	}
	return false
}

func (this *ImapServer) cmdSelect(input string) bool {
	inputN := strings.SplitN(input, " ", 3)

	if len(inputN) == 3 && this.cmdCompare(inputN[1], CMD_SELECT) {
		this.selectBox = strings.Trim(inputN[2], "\"")
		msgCount, _ := models.BoxUserMessageCountByClassName(this.userID, this.selectBox)
		this.writeArgs("* %d EXISTS", msgCount)
		this.writeArgs("* 0 RECENT")
		this.writeArgs("* OK [UIDVALIDITY 1] UIDs valid")
		this.writeArgs("* FLAGS (\\Answered\\Seen \\Deleted \\Draft \\Flagged)")
		this.writeArgs("* OK [PERMANENTFLAGS (\\Answered \\Seen \\Deleted \\Draft \\Flagged)] Limited")
		this.writeArgs("%s OK [READ-WRITE] %s completed", inputN[0], inputN[1])
		return true
	}
	return false
}

func (this *ImapServer) cmdFetch(input string) bool {
	inputN := strings.SplitN(input, " ", 4)

	if len(inputN) == 4 && this.cmdCompare(inputN[1], CMD_FETCH) {
		// fmt.Println("fetch:%s", input)

		list, err := models.BoxAllByClassName(this.userID, this.selectBox)
		fmt.Println(list)
		if err == nil {
			for i := 1; i <= len(list); i++ {
				this.writeArgs("* %d FETCH (UID %s)", i, list[i-1]["mid"].(string))
			}
		}
		this.writeArgs(MSG_COMPLELED, inputN[0], inputN[1])
		return true
	}
	return false
}

func (this *ImapServer) cmdUid(input string) bool {
	inputN := strings.SplitN(input, " ", 5)

	if len(inputN) == 5 && this.cmdCompare(inputN[1], CMD_UID) {

		if strings.EqualFold(inputN[2], "fetch") {
			// this.w("* 1 FETCH (UID 1320476750)\r\n")
			// this.w("* 2 FETCH (UID 1320476751)\r\n")

			list, err := models.BoxAllByClassName(this.userID, this.selectBox)
			if err == nil {
				for i := 1; i <= len(list); i++ {
					c := this.parseArgsConent(inputN[4], list[i-1]["mid"].(string))
					// fmt.Println(c)
					// d := fmt.Sprintf("")
					this.writeArgs("* %d FETCH "+c, i)
					// this.writeArgs("* %d FETCH (UID %s)", i, list[i-1]["mid"].(string))
				}
			}
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
			list, err := models.ClassGetByUid(this.userID)
			if err == nil {
				for i := 1; i <= len(list); i++ {
					fmt.Println(list[i-1]["flags"], list[i-1]["name"])
					mailbox, _ := utf7.Encoding.NewEncoder().String(list[i-1]["name"].(string))
					fmt.Println(mailbox)
					this.writeArgs("* LIST (\\%s) \"/\" \"%s\"", list[i-1]["flags"], mailbox)
				}
				this.writeArgs(MSG_COMPLELED, inputN[0], inputN[1])
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

func (this *ImapServer) cmdNoop(input string) bool {
	inputN := strings.SplitN(input, " ", 2)
	if len(inputN) == 2 {
		if this.cmdCompare(inputN[1], CMD_NOOP) {
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

			if this.cmdNoop(input) {
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
