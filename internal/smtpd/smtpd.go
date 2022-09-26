package smtpd

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"strings"
	"time"

	"github.com/midoks/imail/internal/conf"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/log"
	"github.com/midoks/imail/internal/tools"
	"github.com/midoks/imail/internal/tools/mail"
)

const (
	CMD_READY = iota
	CMD_STARTTLS
	CMD_HELO
	CMD_EHLO
	CMD_AUTH_PLAIN
	CMD_AUTH_LOGIN
	CMD_AUTH_LOGIN_USER
	CMD_AUTH_LOGIN_PWD
	CMD_MAIL_FROM
	CMD_RCPT_TO
	CMD_DATA
	CMD_DATA_END
	CMD_QUIT
)

var stateList = map[int]string{
	CMD_READY:          "READY",
	CMD_STARTTLS:       "STARTTLS",
	CMD_HELO:           "HELO",
	CMD_EHLO:           "EHLO",
	CMD_AUTH_LOGIN:     "AUTH LOGIN",
	CMD_AUTH_LOGIN_PWD: "PASSWORD",
	CMD_AUTH_PLAIN:     "AUTH PLAIN",
	CMD_MAIL_FROM:      "MAIL FROM",
	CMD_RCPT_TO:        "RCPT TO",
	CMD_DATA:           "DATA",
	CMD_DATA_END:       ".",
	CMD_QUIT:           "QUIT",
}

// https://datatracker.ietf.org/doc/html/rfc5321#page-65
var stateTimeout = map[int]int64{
	CMD_READY:          300,
	CMD_STARTTLS:       300,
	CMD_HELO:           300,
	CMD_EHLO:           300,
	CMD_AUTH_LOGIN:     300,
	CMD_AUTH_LOGIN_PWD: 300,
	CMD_AUTH_PLAIN:     300,
	CMD_MAIL_FROM:      300,
	CMD_RCPT_TO:        300,
	CMD_DATA:           120,
	CMD_DATA_END:       180,
	CMD_QUIT:           5,
}

const (
	MSG_INIT            = "220.init"
	MSG_OK              = "250.ok"
	MSG_MAIL_OK         = "250"
	MSG_BYE             = "221"
	MSG_BAD_SYNTAX      = "500"
	MSG_COMMAND_HE_ERR  = "503"
	MSG_COMMAND_ERR     = "502"
	MSG_BAD_USER        = "505"
	MSG_BAD_OPEN_RELAY  = "505.open.relay"
	MSG_BAD_MAIL_ADDR   = "554"
	MSG_COMMAND_TM_ERR  = "421"
	MSG_COMMAND_TM_CTC  = "421.ctc"
	MSG_AUTH_LOGIN_USER = "334.user"
	MSG_AUTH_LOGIN_PWD  = "334.passwd"
	MSG_AUTH_OK         = "235"
	MSG_AUTH_FAIL       = "535"
	MSG_DATA            = "354"
	MSG_STARTTLS        = "220"
)

var msgList = map[string]string{
	MSG_INIT:            "cachecha.com Anti-spam GT for Coremail System (imail[20210626])",
	MSG_OK:              "ok",
	MSG_BYE:             "bye",
	MSG_COMMAND_HE_ERR:  "Error: send HELO/EHLO first",
	MSG_COMMAND_ERR:     "Error: command not implemented",
	MSG_COMMAND_TM_CTC:  "closing transmission channel",
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
	MSG_STARTTLS:        "Ready to start TLS from xxx to mail.xxx.com.",
}

var GO_EOL = tools.GetGoEol()

// Protocol represents the protocol used in the SMTP session
type Protocol string

const (
	// SMTP
	SMTP Protocol = "SMTP"

	// Extended SMTP
	ESMTP = "ESMTP"
)

// Peer represents the client connecting to the server
type Peer struct {
	HeloName         string   // Server name used in HELO/EHLO command
	Username         string   // Username from authentication, if authenticated
	Password         string   // Password from authentication, if authenticated
	Protocol         Protocol // Protocol used, SMTP or ESMTP
	ServerName       string   // A copy of Server.Hostname
	Addr             net.Addr // Network address
	ReceivedMailAddr string   // Received Mail Address
}

type SmtpdServer struct {
	method            int
	isLogin           bool
	conn              net.Conn
	state             int
	startTime         time.Time
	errCount          int
	loginUser         string
	loginPwd          string
	recordCmdMailFrom string
	recordcmdRcptTo   string
	recordCmdData     string

	reader  *bufio.Reader
	writer  *bufio.Writer
	scanner *bufio.Scanner

	//DB DATA
	userID int64

	//run mode
	modeIn bool

	// Determine the current mode of operation
	// open mail in mode
	runModeIn bool

	peer Peer

	//tls
	enableStartTtls bool
	LinkSSL         bool
	tls             bool
	stateTLS        *tls.ConnectionState
	TLSConfig       *tls.Config // Enable STARTTLS support.

	//global
	Domain string
}

func (smtp *SmtpdServer) base64Encode(en string) string {
	src := []byte(en)
	maxLen := base64.StdEncoding.EncodedLen(len(src))
	dst := make([]byte, maxLen)
	base64.StdEncoding.Encode(dst, src)
	return string(dst)
}

func (smtp *SmtpdServer) base64Decode(de string) string {
	dst, err := base64.StdEncoding.DecodeString(de)
	if err != nil {
		return ""
	}
	return string(dst)
}

func (smtp *SmtpdServer) SetReadDeadline(state int) {
	var timeout int64
	timeout = stateTimeout[state]
	smtp.conn.SetReadDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
}

func (smtp *SmtpdServer) setState(state int) {
	smtp.state = state
}

func (smtp *SmtpdServer) getState() int {
	return smtp.state
}

func (smtp *SmtpdServer) D(format string, args ...interface{}) {

	info := fmt.Sprintf(format, args...)
	info = strings.TrimSpace(info)

	if smtp.LinkSSL {
		log.Debugf("[SSL]:%s", info)
		return
	}

	// if conf.Smtp.Debug {
	log.Debug(info)
	// }
}

func (smtp *SmtpdServer) w(msg string) error {
	smtp.D("smtpd[w][%s]:%s", smtp.peer.Addr, msg)

	_, err := smtp.writer.Write([]byte(msg))
	smtp.writer.Flush()
	return err
}

func (smtp *SmtpdServer) write(code string) error {
	info := fmt.Sprintf("%.3s %s%s", code, msgList[code], GO_EOL)
	return smtp.w(info)
}

func (smtp *SmtpdServer) getString(state int) (string, error) {
	if state == CMD_DATA {
		return "", nil
	}
	input, err := smtp.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	inputTrim := strings.TrimSpace(input)
	smtp.D("smtpd[r][%s]:%s", smtp.peer.Addr, inputTrim)
	return inputTrim, err

}

func (smtp *SmtpdServer) close() {
	smtp.conn.Close()
}

func (smtp *SmtpdServer) cmdCompare(input string, cmd int) bool {
	if strings.EqualFold(input, stateList[cmd]) {
		return true
	}
	return false
}

func (smtp *SmtpdServer) stateCompare(input int, cmd int) bool {
	if input == cmd {
		return true
	}
	return false
}

func (smtp *SmtpdServer) cmdHelo(input string) bool {
	inputN := strings.SplitN(input, " ", 2)
	if len(inputN) == 2 {
		if smtp.cmdCompare(inputN[0], CMD_HELO) {
			smtp.peer.HeloName = inputN[1]

			smtp.D("smtpd[helo]:%s", inputN[1])
			smtp.write(MSG_OK)
			return true
		}
	}
	return false
}

func (smtp *SmtpdServer) cmdEhlo(input string) bool {
	inputN := strings.SplitN(input, " ", 2)
	if len(inputN) == 2 {
		if smtp.cmdCompare(inputN[0], CMD_EHLO) {
			smtp.w(fmt.Sprintf("250-mail%s", GO_EOL))
			smtp.w(fmt.Sprintf("250-PIPELINING%s", GO_EOL))
			smtp.w(fmt.Sprintf("250-AUTH LOGIN PLAIN%s", GO_EOL))
			smtp.w(fmt.Sprintf("250-AUTH=LOGIN PLAIN%s", GO_EOL))
			smtp.w(fmt.Sprintf("250-coremail 1Uxr2xKj7kG0xkI17xGrU7I0s8FY2U3Uj8Cz28x1UUUUU7Ic2I0Y2UFRbmXhUCa0xDrUUUUj%s", GO_EOL))
			if smtp.enableStartTtls {
				smtp.w(fmt.Sprintf("250-STARTTLS%s", GO_EOL))
			}

			smtp.w(fmt.Sprintf("250-SIZE 73400320%s", GO_EOL))
			smtp.w(fmt.Sprintf("250 8BITMIME%s", GO_EOL))
			return true
		}
	}
	return false
}

func (smtp *SmtpdServer) cmdAuthLogin(input string) bool {
	if smtp.cmdCompare(input, CMD_AUTH_LOGIN) {
		smtp.write(MSG_AUTH_LOGIN_USER)
		return true
	}
	// smtp.write(MSG_BAD_SYNTAX)
	return false
}

func (smtp *SmtpdServer) checkUserLogin() bool {
	name := smtp.loginUser
	pwd := strings.TrimSpace(smtp.loginPwd)

	isLogin, id := db.LoginWithCode(name, pwd)

	if !isLogin {
		return false
	}

	smtp.userID = id
	smtp.isLogin = true
	return true
}

func (smtp *SmtpdServer) cmdAuthLoginUser(input string) bool {
	user := smtp.base64Decode(input)
	smtp.loginUser = user

	smtp.D("smtpd[user]:%s", smtp.loginUser)
	smtp.write(MSG_AUTH_LOGIN_PWD)
	return true
}

func (smtp *SmtpdServer) cmdAuthLoginPwd(input string) bool {

	pwd := smtp.base64Decode(input)
	smtp.loginPwd = pwd

	if smtp.checkUserLogin() {
		smtp.write(MSG_AUTH_OK)
		return true
	}
	smtp.write(MSG_AUTH_FAIL)

	//fail log to db
	info := fmt.Sprintf("[smtp]user[%s]:%.3s %s%s", smtp.loginUser, MSG_AUTH_FAIL, msgList[MSG_AUTH_FAIL], GO_EOL)
	db.LogAdd("auth_plain_login", info)
	return false
}

func (smtp *SmtpdServer) cmdAuthPlainLogin(input string) (bool, bool) {
	if strings.HasPrefix(input, stateList[CMD_AUTH_PLAIN]) {
		inputN := strings.SplitN(input, " ", 3)
		if len(inputN) == 3 {
			data := smtp.base64Decode(inputN[2])

			list := strings.SplitN(data, "\x00", 3)

			userList := strings.Split(list[1], "@")

			smtp.loginUser = userList[0]
			smtp.loginPwd = list[2]

			b := smtp.checkUserLogin()
			if b {
				smtp.write(MSG_AUTH_OK)
				return true, true
			}
			smtp.write(MSG_AUTH_FAIL)

			//fail log to db
			info := fmt.Sprintf("[smtp]user[%s]:%.3s %s%s", smtp.loginUser, MSG_AUTH_FAIL, msgList[MSG_AUTH_FAIL], GO_EOL)
			db.LogAdd("auth_plain_login", info)

			return true, false
		}

	}
	return false, false
}

func (smtp *SmtpdServer) isAllowDomain(domain string) bool {
	return db.DomainVaild(domain)
}

func (smtp *SmtpdServer) cmdMailFrom(input string) bool {
	inputN := strings.SplitN(input, ":", 2)

	if len(inputN) == 2 && smtp.cmdCompare(inputN[0], CMD_MAIL_FROM) {

		inputN[1] = strings.TrimSpace(inputN[1])
		inputN[1] = tools.FilterAddressBody(inputN[1])

		if !tools.CheckStandardMail(inputN[1]) {
			smtp.write(MSG_BAD_SYNTAX)
			return false
		}

		mailFrom := tools.GetRealMail(inputN[1])
		if !tools.IsEmailRe(mailFrom) {
			smtp.write(MSG_BAD_USER)
			return false
		}

		if smtp.isLogin {
			info := strings.Split(mailFrom, "@")

			if !smtp.isAllowDomain(info[1]) {
				smtp.write(MSG_BAD_MAIL_ADDR)
				return false
			}

			user, err := db.UserGetByName(info[0])
			if err != nil {
				smtp.write(MSG_BAD_USER)
				return false
			}
			smtp.userID = user.Id
		}

		smtp.recordCmdMailFrom = mailFrom
		smtp.write(MSG_MAIL_OK)

		return true
	}
	return false
}

func (smtp *SmtpdServer) cmdModeInMailFrom(input string) bool {
	inputN := strings.SplitN(input, ":", 2)

	if len(inputN) == 2 {
		if smtp.cmdCompare(inputN[0], CMD_MAIL_FROM) {

			inputN[1] = strings.TrimSpace(inputN[1])
			inputN[1] = tools.FilterAddressBody(inputN[1])

			if !tools.CheckStandardMail(inputN[1]) {
				smtp.write(MSG_BAD_SYNTAX)
				return false
			}

			mailFrom := tools.GetRealMail(inputN[1])
			if !tools.IsEmailRe(mailFrom) {
				smtp.write(MSG_BAD_USER)
				return false
			}

			smtp.recordCmdMailFrom = mailFrom
			smtp.write(MSG_MAIL_OK)

			return true
		}
	}
	return false
}

func (smtp *SmtpdServer) cmdStartTtls(input string) bool {
	if smtp.tls {
		smtp.write(MSG_STARTTLS)
		return true
	}

	smtp.initTLSConfig()

	if smtp.TLSConfig == nil {
		smtp.w("502 Error: TLS not supported")
		return false
	}

	tlsConn := tls.Server(smtp.conn, smtp.TLSConfig)
	smtp.w("220 Go ahead\n")

	if err := tlsConn.Handshake(); err != nil {
		errmsg := fmt.Sprintf("550 ERROR: Handshake error:%s", err)
		smtp.w(errmsg)
		return false
	}

	state := tlsConn.ConnectionState()

	smtp.reader = bufio.NewReader(tlsConn)
	smtp.writer = bufio.NewWriter(tlsConn)
	smtp.scanner = bufio.NewScanner(smtp.reader)

	smtp.stateTLS = &state
	smtp.tls = true
	return true
}

func (smtp *SmtpdServer) cmdRcptTo(input string) bool {
	inputN := strings.SplitN(input, ":", 2)

	if len(inputN) == 2 {
		if smtp.cmdCompare(inputN[0], CMD_RCPT_TO) {
			inputN[1] = strings.TrimSpace(inputN[1])

			if !tools.CheckStandardMail(inputN[1]) {
				smtp.write(MSG_BAD_SYNTAX)
				return false
			}

			rcptTo := tools.GetRealMail(inputN[1])

			if !tools.IsEmailRe(rcptTo) {
				smtp.write(MSG_BAD_USER)
				return false
			}
			smtp.recordcmdRcptTo = rcptTo

			if smtp.runModeIn { //外部邮件,邮件地址检查
				info := strings.Split(rcptTo, "@")

				if !smtp.isAllowDomain(info[1]) {
					smtp.write(MSG_BAD_OPEN_RELAY)
					return false
				}
				user, err := db.UserGetByName(info[0])
				if err != nil {
					smtp.write(MSG_BAD_USER)
					return false
				}
				smtp.userID = user.Id
			}

			smtp.write(MSG_MAIL_OK)
			return true
		}
	}
	smtp.write(MSG_BAD_SYNTAX)
	return false
}

func (smtp *SmtpdServer) cmdData(input string) bool {
	if smtp.cmdCompare(input, CMD_DATA) {
		smtp.write(MSG_DATA)
		return true
	}
	smtp.write(MSG_BAD_SYNTAX)
	return false
}

func (smtp *SmtpdServer) addEnvelopeDataAcceptLine(data []byte) []byte {
	tlsDetails := ""

	tlsVersions := map[uint16]string{
		tls.VersionSSL30: "SSL3.0",
		tls.VersionTLS10: "TLS1.0",
		tls.VersionTLS11: "TLS1.1",
		tls.VersionTLS12: "TLS1.2",
		tls.VersionTLS13: "TLS1.3",
	}

	if smtp.stateTLS != nil {
		version := "unknown"

		if val, ok := tlsVersions[smtp.stateTLS.Version]; ok {
			version = val
		}

		cipher := tls.CipherSuiteName(smtp.stateTLS.CipherSuite)

		tlsDetails = fmt.Sprintf(
			"\r\n\t(version=%s cipher=%s);",
			version,
			cipher,
		)
	}

	peerIP := ""
	if addr, ok := smtp.peer.Addr.(*net.TCPAddr); ok {
		peerIP = addr.IP.String()
	}

	serverTagName := fmt.Sprintf("smtp.%s (NewMx)", smtp.Domain)

	line := tools.Wrap([]byte(fmt.Sprintf(
		"Received: from %s (unknown[%s])\n\tby %s with SMTP id\n\tfor <%s>; %s %s\r\n",
		peerIP,
		peerIP,
		serverTagName,
		smtp.recordCmdMailFrom,
		time.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700 (MST)"),
		tlsDetails,
	)))

	data = append(data, line...)

	// Move the new Received line up front
	copy(data[len(line):], data[0:len(data)-len(line)])
	copy(data, line)
	return data

}

func (smtp *SmtpdServer) cmdDataAccept() bool {

	data := &bytes.Buffer{}
	reader := textproto.NewReader(smtp.reader).DotReader()
	_, err := io.CopyN(data, reader, int64(10240000))

	content := string(data.Bytes())
	if err == io.EOF {
		smtp.write(MSG_MAIL_OK)
	}
	// smtp.D("smtpd[data]:", content)

	if smtp.runModeIn {
		// smtp.D("smtpd[data][peer]:", smtp.peer)
		revContent := string(smtp.addEnvelopeDataAcceptLine(data.Bytes()))
		fid, err := db.MailPush(smtp.userID, 1, smtp.recordCmdMailFrom, smtp.recordcmdRcptTo, revContent, 3, false)
		if err != nil {
			return false
		}
		mail.ExecPython(conf.Hook.SendScript, fid)
	} else {
		fid, err := db.MailPush(smtp.userID, 0, smtp.recordCmdMailFrom, smtp.recordcmdRcptTo, content, 0, false)
		if err != nil {
			return false
		}
		mail.ExecPython(conf.Hook.ReceiveScript, fid)
	}
	return true
}

func (smtp *SmtpdServer) cmdQuit(input string) bool {
	if smtp.cmdCompare(input, CMD_QUIT) {
		smtp.write(MSG_BYE)
		smtp.close()
		return true
	}
	return false
}

// 本地用户邮件投递到其他邮件地址|需要登陆
func (smtp *SmtpdServer) cmdModeOut(state int, input string) bool {
	//CMD_AUTH_LOGIN
	if smtp.stateCompare(state, CMD_AUTH_LOGIN) {
		if smtp.cmdAuthLoginUser(input) {
			smtp.setState(CMD_AUTH_LOGIN_USER)
		}
	}

	//CMD_AUTH_LOGIN_USER
	if smtp.stateCompare(state, CMD_AUTH_LOGIN_USER) {
		if smtp.cmdQuit(input) {
			return true
		}

		if smtp.cmdAuthLoginPwd(input) {
			smtp.setState(CMD_AUTH_LOGIN_PWD)
		}
	}

	//CMD_AUTH_LOGIN_PWD
	if smtp.stateCompare(state, CMD_AUTH_LOGIN_PWD) {
		if smtp.cmdQuit(input) {
			return true
		}

		if smtp.cmdMailFrom(input) {
			smtp.setState(CMD_MAIL_FROM)
		}
	}

	return false
}

func (smtp *SmtpdServer) handle() {
	for {

		state := smtp.getState()

		input, err := smtp.getString(state)

		if err != nil {
			smtp.D("smtp: %s", err)
			smtp.write(MSG_COMMAND_TM_CTC)
			smtp.close()
			break
		}

		if smtp.cmdQuit(input) {
			break
		}

		//CMD_READY
		if smtp.stateCompare(state, CMD_READY) {
			smtp.runModeIn = false
			if smtp.cmdHelo(input) {
				smtp.setState(CMD_HELO)
			} else if smtp.cmdEhlo(input) {
				smtp.setState(CMD_EHLO)
			} else {
				smtp.write(MSG_COMMAND_HE_ERR)
			}
		}

		//CMD_HELO
		if smtp.stateCompare(state, CMD_HELO) || smtp.stateCompare(state, CMD_EHLO) {

			if smtp.cmdHelo(input) {
				smtp.setState(CMD_HELO)
			} else if smtp.cmdEhlo(input) {
				smtp.setState(CMD_EHLO)
			}

			if smtp.modeIn {
				if smtp.cmdModeInMailFrom(input) {
					smtp.setState(CMD_MAIL_FROM)
					smtp.runModeIn = true
				}
			}
		}

		if smtp.enableStartTtls { //CMD_STARTTLS
			if input == stateList[CMD_STARTTLS] {
				if !smtp.cmdStartTtls(input) {
					break
				}
			}
		}

		if smtp.runModeIn {
			//收取外邮模式
			//CMD_MAIL_FROM
			if smtp.stateCompare(state, CMD_MAIL_FROM) {

				if smtp.cmdRcptTo(input) {
					smtp.setState(CMD_RCPT_TO)
				}
			}

			//CMD_RCPT_TO
			if smtp.stateCompare(state, CMD_RCPT_TO) {

				if smtp.cmdData(input) {
					smtp.setState(CMD_DATA)
				}
			}

		} else {
			//登录模式
			isAuthPlain, isAuthPlainOK := smtp.cmdAuthPlainLogin(input)

			if isAuthPlain && isAuthPlainOK {
				smtp.setState(CMD_AUTH_LOGIN_PWD)
			} else if isAuthPlain && !isAuthPlainOK { //AUTH PLAIN FAIL
				break
			} else if smtp.cmdAuthLogin(input) {
				smtp.setState(CMD_AUTH_LOGIN)
			}

			isBreak := smtp.cmdModeOut(state, input)
			if isBreak {
				break
			}

			//CMD_MAIL_FROM
			if smtp.stateCompare(state, CMD_MAIL_FROM) {

				if smtp.cmdRcptTo(input) {
					smtp.setState(CMD_RCPT_TO)
				}
			}

			//CMD_RCPT_TO
			if smtp.stateCompare(state, CMD_RCPT_TO) {

				if smtp.cmdData(input) {
					smtp.setState(CMD_DATA)
				}
			}
		}

		//CMD_DATA
		if smtp.stateCompare(state, CMD_DATA) {
			if smtp.cmdDataAccept() {
				smtp.setState(CMD_DATA_END)
			}
		}

		//CMD_DATA_END
		if smtp.stateCompare(state, CMD_DATA_END) {
			smtp.setState(CMD_READY)
		}
	}
}

func (smtp *SmtpdServer) initTLSConfig() {
	smtp.TLSConfig = tools.InitAutoMakeTLSConfig()
}

func (smtp *SmtpdServer) ready() {

	if smtp.LinkSSL {
		smtp.initTLSConfig()
	}

	smtp.startTime = time.Now()
	smtp.isLogin = false
	smtp.enableStartTtls = true

	//mode
	smtp.runModeIn = false
	smtp.modeIn = conf.Smtp.ModeIn
	smtp.Domain = conf.Web.Domain
}

func (smtp *SmtpdServer) start(conn net.Conn) {
	smtp.conn = conn

	smtp.reader = bufio.NewReader(conn)
	smtp.writer = bufio.NewWriter(conn)
	smtp.scanner = bufio.NewScanner(smtp.reader)

	defer conn.Close()

	if smtp.enableStartTtls {
		var tlsConn *tls.Conn
		if tlsConn, smtp.tls = conn.(*tls.Conn); smtp.tls {
			tlsConn.Handshake()
			tlsState := tlsConn.ConnectionState()
			smtp.stateTLS = &tlsState
		}
	}

	smtp.peer = Peer{
		Addr: conn.RemoteAddr(),
		// ServerName: conn.Hostname,
	}

	smtp.write(MSG_INIT)
	smtp.setState(CMD_READY)
	smtp.SetReadDeadline(CMD_READY)
	smtp.handle()

}

func (smtp *SmtpdServer) StartPort(port int) {
	smtp.ready()
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		smtp.D("StartPort:%s", err)
		return
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go smtp.start(conn)
	}
}

func (smtp *SmtpdServer) StartSSLPort(port int) {
	smtp.LinkSSL = true
	smtp.ready()
	addr := fmt.Sprintf(":%d", port)
	ln, err := tls.Listen("tcp", addr, smtp.TLSConfig)
	if err != nil {
		smtp.D("[smtp]StartSSLPort:%s", err)
		return
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go smtp.start(conn)
	}
}

func Start(port int) {
	srv := SmtpdServer{}
	srv.StartPort(port)
}

func StartSSL(port int) {
	srvSSL := SmtpdServer{}
	srvSSL.StartSSLPort(port)
}
