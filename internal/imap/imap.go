package imap

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/midoks/imail/internal/config"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/imap/component"
	"github.com/midoks/imail/internal/libs"
	"github.com/midoks/imail/internal/log"
	"io"
	"net"
	"net/textproto"
	"strconv"
	"strings"
	"time"
)

const (
	CMD_READY = iota
	CMD_AUTH
	CMD_LIST
	CMD_XLIST
	CMD_LOGOUT
	CMD_CAPABILITY
	CMD_ID
	CMD_STATUS
	CMD_SELECT
	CMD_FETCH
	CMD_APPEND
	CMD_UID
	CMD_COPY
	CMD_CLOSE
	CMD_STORE
	CMD_NAMESPACE
	CMD_SEARCH
	CMD_NOOP
	CMD_EXPUNGE
)

var stateList = map[int]string{
	CMD_READY:      "READY",
	CMD_AUTH:       "LOGIN",
	CMD_LOGOUT:     "LOGOUT",
	CMD_LIST:       "LIST",
	CMD_XLIST:      "XLIST",
	CMD_CAPABILITY: "CAPABILITY",
	CMD_ID:         "ID",
	CMD_STATUS:     "STATUS",
	CMD_SELECT:     "SELECT",
	CMD_FETCH:      "FETCH",
	CMD_COPY:       "COPY",
	CMD_APPEND:     "APPEND",
	CMD_STORE:      "STORE",
	CMD_NAMESPACE:  "NAMESPACE",
	CMD_SEARCH:     "SEARCH",
	CMD_UID:        "UID",
	CMD_NOOP:       "NOOP",
	CMD_CLOSE:      "CLOSE",
	CMD_EXPUNGE:    "EXPUNGE",
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

// var GO_EOL = "\n"
var GO_EOL = libs.GetGoEol()

// https://datatracker.ietf.org/doc/html/rfc3501#page-48
type UIDVNW struct {
	Copy    bool
	Expunge bool
}

type ImapServer struct {
	io.Reader
	io.RuneScanner
	// *Writer
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

	TLSConfig *tls.Config // Enable STARTTLS support.

	//Turn off function related
	nl        net.Listener
	nlConn    net.Conn
	nlSSL     net.Listener
	nlConnSSL net.Conn

	uidVnw UIDVNW
}

func (this *ImapServer) setState(state int) {
	this.state = state
}

func (this *ImapServer) getState() int {
	return this.state
}

func (this *ImapServer) D(args ...interface{}) {

	imapDebug, _ := config.GetBool("imap.debug", false)
	if imapDebug {
		// fmt.Println(args...)
		log.Info(args...)
	}
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

func (this *ImapServer) parseArgsConent(format string, data db.Mail) (string, error) {

	content := data.Content
	id := data.Id

	format = strings.TrimSpace(format)
	format = strings.Trim(format, "()")

	inputN := strings.Split(format, " ")
	list := make(map[string]interface{})

	bufferedBody := bufio.NewReader(strings.NewReader(content))
	header, err := component.ReadHeader(bufferedBody)

	if err != nil {
		return "", err
	}

	bs, err := component.FetchBodyStructure(header, bufferedBody, true)

	for i := 0; i < len(inputN); i++ {

		if strings.EqualFold(inputN[i], "uid") {
			uid_id := fmt.Sprintf("%d", id)
			list[inputN[i]] = uid_id
		}

		if strings.EqualFold(inputN[i], "flags") {
			flags := "("
			if data.IsRead > 0 {
				flags += "\\Seen"
			} else {
				flags += "\\UNSEEN"
			}

			if data.IsFlags > 0 {
				flags += "\\Flagged"
			}

			flags += ")"
			list[inputN[i]] = flags
		}

		if strings.EqualFold(inputN[i], "rfc822.size") {
			rfc822_size := fmt.Sprintf("%d", len(content))
			list[inputN[i]] = rfc822_size
		}

		if strings.EqualFold(inputN[i], "bodystructure") {
			list[inputN[i]] = bs.ToString()
		}

		if strings.EqualFold(inputN[i], "body.peek[header]") {
			headerString, _ := component.ReadHeaderString(bufferedBody)
			list["body[header]"] = fmt.Sprintf("{%d}\r\n%s", len(headerString), headerString)
		}

		if strings.EqualFold(inputN[i], "body.peek[]") {
			list["body[]"] = fmt.Sprintf("{%d}\r\n%s", len(content), content)
			db.MailSeenById(id)
		}
	}

	out := ""
	for i := 0; i < len(inputN); i++ {
		if strings.EqualFold(inputN[i], "body.peek[header]") {
			out += fmt.Sprintf("%s %s ", strings.ToUpper("body[header]"), list["body[header]"])
		} else if strings.EqualFold(inputN[i], "body.peek[]") {
			out += fmt.Sprintf("%s %s", strings.ToUpper("body[]"), list["body[]"])
		} else {
			out += fmt.Sprintf("%s %s ", strings.ToUpper(inputN[i]), list[inputN[i]])
		}
	}

	out = fmt.Sprintf("(%s)", out)
	return out, nil
}

func (this *ImapServer) cmdAuth(input string) bool {
	inputN := strings.SplitN(input, " ", 4)
	if len(inputN) == 4 && this.cmdCompare(inputN[1], CMD_AUTH) {
		if len(inputN) < 4 {
			this.writeArgs(MSG_BAD_SYNTAX, inputN[0])
			return false
		}

		user := strings.Trim(inputN[2], "\"")
		pwd := strings.Trim(inputN[3], "\"")

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

func (this *ImapServer) cmdNameSpace(input string) bool {

	inputN := strings.SplitN(input, " ", 2)
	if len(inputN) == 2 {
		if this.cmdCompare(inputN[1], CMD_NAMESPACE) {
			this.writeArgs("* NAMESPACE ((\"\" \"/\")) NIL NIL")
			this.writeArgs(MSG_COMPLELED, inputN[0], inputN[1])
			return true
		}
	}
	return false
}

func (this *ImapServer) cmdList(input string) bool {
	inputN := strings.SplitN(input, " ", 4)
	if len(inputN) == 4 {
		if this.cmdCompare(inputN[1], CMD_LIST) || this.cmdCompare(inputN[1], CMD_XLIST) {
			this.writeArgs("* %s (\\NoSelect \\HasChildren) \"/\" \"&UXZO1mWHTvZZOQ-\"", inputN[1])
			this.writeArgs("* %s (\\HasChildren) \"/\" \"INBOX\"", inputN[1])
			this.writeArgs("* %s (\\HasChildren) \"/\" \"Sent Messages\"", inputN[1])
			this.writeArgs("* %s (\\HasChildren) \"/\" \"Drafts\"", inputN[1])
			this.writeArgs("* %s (\\HasChildren) \"/\" \"Deleted Messages\"", inputN[1])
			this.writeArgs("* %s (\\HasChildren) \"/\" \"Junk\"", inputN[1])
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
			this.writeArgs("* FLAGS (\\Answered \\Seen \\Deleted \\Draft)")
			this.writeArgs("* OK [PERMANENTFLAGS (\\Answered \\Seen \\Deleted \\Draft)] Limited")
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
			mailList := db.MailListForImap(this.userID)
			for i, m := range mailList {
				this.writeArgs("* %d FETCH (UID %d)", i+1, m.Id)
			}

			this.writeArgs(MSG_COMPLELED, inputN[0], inputN[1])
			return true
		}
	}
	return false
}

func (this *ImapServer) cmdAppend(input string) bool {
	inputN := strings.SplitN(input, " ", 5)
	if len(inputN) == 5 {
		if this.cmdCompare(inputN[1], CMD_APPEND) {
			this.w("+ Ready for literal data")

			data := &bytes.Buffer{}
			reader := textproto.NewReader(this.reader).DotReader()
			_, err := io.CopyN(data, reader, int64(10240000))
			content := string(data.Bytes())
			fmt.Println(content, err)
			this.writeArgs("%s OK %s completed", inputN[0], inputN[1])
			return true
		}
	}
	return false
}

// https://datatracker.ietf.org/doc/html/rfc3501#page-48
func (this *ImapServer) cmdUid(input string) bool {

	inputN := strings.SplitN(input, " ", 5)
	if len(inputN) == 5 {
		if this.cmdCompare(inputN[1], CMD_UID) {
			if this.cmdCompare(inputN[2], CMD_FETCH) {

				if strings.Index(inputN[3], ":") > 0 {
					se := strings.SplitN(inputN[3], ":", 2)
					start, _ := strconv.ParseInt(se[0], 10, 64)
					end, _ := strconv.ParseInt(se[1], 10, 64)
					mailList, _ := db.BoxListByImap(this.userID, this.selectBox, start, end)
					for i, m := range mailList {

						c, _ := this.parseArgsConent(inputN[4], m)
						this.writeArgs("* %d FETCH %s", i+1, c)
					}
				}

				if libs.IsNumeric(inputN[3]) {
					mid, _ := strconv.ParseInt(inputN[3], 10, 64)
					mailList, _ := db.BoxListByMid(this.userID, this.selectBox, mid)
					c, _ := this.parseArgsConent(inputN[4], mailList[0])
					this.writeArgs("* %d FETCH %s", mid, c)
				}
			} else if this.cmdCompare(inputN[2], CMD_SEARCH) {

				if strings.Index(inputN[4], ":") > 0 {
					se := strings.SplitN(inputN[4], ":", 2)
					start, _ := strconv.ParseInt(se[0], 10, 64)
					end, _ := strconv.ParseInt(se[1], 10, 64)
					mailList, _ := db.BoxListByImap(this.userID, this.selectBox, start, end)
					idString := ""
					for _, m := range mailList {
						idString += fmt.Sprintf(" %d", m.Id)
					}
					this.writeArgs("* SEARCH%s", idString)
				}

				if libs.IsNumeric(inputN[3]) {
					mid, _ := strconv.ParseInt(inputN[3], 10, 64)
					mailList, _ := db.BoxListByMid(this.userID, this.selectBox, mid)
					c, _ := this.parseArgsConent(inputN[4], mailList[0])
					this.writeArgs("* %d SEARCH "+c, mid)
				}
			} else if this.cmdCompare(inputN[2], CMD_COPY) {

				if libs.IsNumeric(inputN[3]) {
					mid, _ := strconv.ParseInt(inputN[3], 10, 64)
					inputN[4] = strings.Trim(inputN[4], "\"")
					if strings.EqualFold(inputN[4], "Deleted Messages") {
						db.MailSoftDeleteById(mid, 1)
						db.MailSetJunkById(mid, 0)
					} else if strings.EqualFold(inputN[4], "INBOX") {
						db.MailSoftDeleteById(mid, 0)
						db.MailSetJunkById(mid, 0)
					} else if strings.EqualFold(inputN[4], "Junk") {
						db.MailSoftDeleteById(mid, 0)
						db.MailSetJunkById(mid, 1)
					}

					this.uidVnw.Copy = true
				}
			} else if this.cmdCompare(inputN[2], CMD_STORE) {

				inputN := strings.SplitN(input, " ", 6)
				if libs.IsNumeric(inputN[3]) {
					mid, _ := strconv.ParseInt(inputN[3], 10, 64)
					inputN[5] = strings.Trim(inputN[5], "()")
					inputN[5] = strings.Trim(inputN[5], "\\")
					if strings.EqualFold(inputN[5], "Seen") && strings.HasPrefix(inputN[4], "+") {
						db.MailSeenById(mid)
					} else if strings.EqualFold(inputN[5], "Seen") && strings.HasPrefix(inputN[4], "-") {
						db.MailUnSeenById(mid)
					}

					if strings.EqualFold(inputN[5], "FLAGGED") && strings.HasPrefix(inputN[4], "+") {
						db.MailSetFlagsById(mid, 1)
					} else if strings.EqualFold(inputN[5], "FLAGGED") && strings.HasPrefix(inputN[4], "-") {
						db.MailSetFlagsById(mid, 0)
					}

					if strings.EqualFold(inputN[5], "DELETED") &&
						strings.HasPrefix(inputN[4], "+") && !this.uidVnw.Copy {
						this.uidVnw.Copy = false
						db.MailSoftDeleteById(mid, 1)
					}
				}
			}

			this.writeArgs("%s OK %s %s completed", inputN[0], inputN[1], inputN[2])
			return true
		}
	}
	return false
}

func (this *ImapServer) cmdExpunge(input string) bool {
	inputN := strings.SplitN(input, " ", 2)
	if len(inputN) == 2 {
		if this.cmdCompare(inputN[1], CMD_EXPUNGE) {
			mailList, _ := db.MailDeletedListAllForImap(this.userID)
			for _, m := range mailList {
				this.writeArgs("* %d EXPUNGE", m.Id)
			}
			this.writeArgs("%s OK %s completed", inputN[0], inputN[1])
			return true
		}
	}
	return false
}

func (this *ImapServer) cmdClose(input string) bool {
	inputN := strings.SplitN(input, " ", 2)
	if len(inputN) == 2 {
		if this.cmdCompare(inputN[1], CMD_CLOSE) {
			this.writeArgs("%s OK %s Completed", inputN[0], inputN[1])
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

		if err != nil {
			this.close()
			break
		}

		if this.cmdCapabitity(input) {
		} else if this.cmdId(input) {
		} else if this.cmdNoop(input) {
		} else if this.cmdAuth(input) {
			this.setState(CMD_AUTH)
		}

		if this.stateCompare(state, CMD_AUTH) {

			if this.cmdNameSpace(input) {
			} else if this.cmdList(input) {
			} else if this.cmdStatus(input) {
			} else if this.cmdSelect(input) {
			} else if this.cmdFecth(input) {
			} else if this.cmdUid(input) {
			} else if this.cmdExpunge(input) {
			} else if this.cmdClose(input) {
				this.close()
				break

			} else if this.cmdLogout(input) {
				this.close()
				break
			}
		}

	}
}

func (this *ImapServer) initTLSConfig() {
	this.TLSConfig = libs.InitAutoMakeTLSConfig()
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

	this.handle()
}

func (this *ImapServer) StartPort(port int) {
	var err error
	addr := fmt.Sprintf(":%d", port)
	this.nl, err = net.Listen("tcp", addr)
	if err != nil {
		this.D("[imap]StartSSLPort:", err)
		return
	}

	defer this.nl.Close()

	for {
		this.nlConn, err = this.nl.Accept()
		if err != nil {
			this.D("imap[StartPort][conn]", err)
			return
		} else {
			this.start(this.nlConn)
		}
	}
}

func (this *ImapServer) StartSSLPort(port int) {
	var err error
	this.initTLSConfig()

	addr := fmt.Sprintf(":%d", port)
	this.nlSSL, err = tls.Listen("tcp", addr, this.TLSConfig)
	if err != nil {
		this.D("imap[StartSSLPort]", err)
		return
	}
	defer this.nlSSL.Close()

	this.nlConnSSL, err = this.nlSSL.Accept()
	if err != nil {
		this.D("imap[StartSSLPort][conn]", err)
		return
	}
	this.start(this.nlConnSSL)

}

func (this *ImapServer) Close() error {
	var err error

	err = this.nl.Close()
	if err != nil {
		return err
	}

	if this.nlConn != nil {
		err = this.nlConn.Close()
		return err
	}

	err = this.nlSSL.Close()
	if err != nil {
		return err
	}

	if this.nlConnSSL != nil {
		err = this.nlConnSSL.Close()
		return err
	}

	return nil
}

var srv ImapServer

func init() {
	srv = ImapServer{}
}

func Close() error {
	return srv.Close()
}

func Start(port int) {
	srv.StartPort(port)
}

func StartSSL(port int) {
	srv.StartSSLPort(port)
}
