package smtpd

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"github.com/midoks/imail/internal/config"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/libs"
	"github.com/midoks/imail/internal/log"
	"io"
	"net"
	"net/textproto"
	"strings"
	"time"
)

const (
	CMD_READY           = iota
	CMD_STARTTLS        = iota
	CMD_HELO            = iota
	CMD_EHLO            = iota
	CMD_AUTH_PLAIN      = iota
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

var GO_EOL = libs.GetGoEol()

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
	debug             bool
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

func (this *SmtpdServer) D(args ...interface{}) {
	if this.LinkSSL {
		log.Debugf("[SSL]:%s", args...)
		return
	}

	smtp3Debug, _ := config.GetBool("smtpd.debug", false)
	if smtp3Debug {
		fmt.Println(args...)
		log.Debug(args...)
	}
}

func (this *SmtpdServer) Debug(d bool) {
	this.debug = d
}

func (this *SmtpdServer) w(msg string) error {
	log := fmt.Sprintf("smtpd[w][%s]:%s", this.peer.Addr, msg)
	this.D(log)

	_, err := this.writer.Write([]byte(msg))
	this.writer.Flush()
	return err
}

func (this *SmtpdServer) write(code string) error {
	info := fmt.Sprintf("%.3s %s%s", code, msgList[code], GO_EOL)
	return this.w(info)
}

func (this *SmtpdServer) getString(state int) (string, error) {
	if state == CMD_DATA {
		return "", nil
	}
	input, err := this.reader.ReadString('\n')
	inputTrim := strings.TrimSpace(input)
	this.D("smtpd[r][", this.peer.Addr, "]:", inputTrim, ":", err)
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
	if len(inputN) == 2 {
		if this.cmdCompare(inputN[0], CMD_HELO) {
			this.peer.HeloName = inputN[1]

			this.D("smtpd[helo]:", inputN[1])
			this.write(MSG_OK)
			return true
		}
	}
	return false
}

func (this *SmtpdServer) cmdEhlo(input string) bool {
	inputN := strings.SplitN(input, " ", 2)
	if len(inputN) == 2 {
		if this.cmdCompare(inputN[0], CMD_EHLO) {
			this.w(fmt.Sprintf("250-mail%s", GO_EOL))
			this.w(fmt.Sprintf("250-PIPELINING%s", GO_EOL))
			this.w(fmt.Sprintf("250-AUTH LOGIN PLAIN%s", GO_EOL))
			this.w(fmt.Sprintf("250-AUTH=LOGIN PLAIN%s", GO_EOL))
			this.w(fmt.Sprintf("250-coremail 1Uxr2xKj7kG0xkI17xGrU7I0s8FY2U3Uj8Cz28x1UUUUU7Ic2I0Y2UFRbmXhUCa0xDrUUUUj%s", GO_EOL))
			if this.enableStartTtls {
				this.w(fmt.Sprintf("250-STARTTLS%s", GO_EOL))
			}

			this.w(fmt.Sprintf("250-SIZE 73400320%s", GO_EOL))
			this.w(fmt.Sprintf("250 8BITMIME%s", GO_EOL))
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
	// this.write(MSG_BAD_SYNTAX)
	return false
}

func (this *SmtpdServer) checkUserLogin() bool {
	name := this.loginUser
	pwd := strings.TrimSpace(this.loginPwd)

	isLogin, id := db.LoginWithCode(name, pwd)

	if !isLogin {
		return false
	}

	this.userID = id
	this.isLogin = true
	return true
}

func (this *SmtpdServer) cmdAuthLoginUser(input string) bool {
	user := this.base64Decode(input)
	this.loginUser = user

	this.D("smtpd:", this.loginUser)
	this.write(MSG_AUTH_LOGIN_PWD)
	return true
}

func (this *SmtpdServer) cmdAuthLoginPwd(input string) bool {

	pwd := this.base64Decode(input)
	this.loginPwd = pwd

	this.D("smtpd:", this.loginPwd)
	if this.checkUserLogin() {
		this.write(MSG_AUTH_OK)
		return true
	}
	this.write(MSG_AUTH_FAIL)
	return false
}

func (this *SmtpdServer) cmdAuthPlainLogin(input string) bool {
	if strings.HasPrefix(input, stateList[CMD_AUTH_PLAIN]) {
		inputN := strings.SplitN(input, " ", 3)
		if len(inputN) == 3 {
			data := this.base64Decode(inputN[2])

			// mdomain := config.GetString("mail.domain", "xxx.com")
			list := strings.SplitN(data, "\x00", 3)
			userList := strings.Split(list[1], "@")

			this.loginUser = userList[0]
			this.loginPwd = list[2]

			b := this.checkUserLogin()
			this.D("smtpd:", b, this.loginUser, this.loginPwd)
			if b {
				this.write(MSG_AUTH_OK)
				return true
			}
			this.write(MSG_AUTH_FAIL)
		}
	}
	return false
}

func (this *SmtpdServer) isAllowDomain(domain string) bool {
	mdomain := config.GetString("mail.domain", "xxx.com")
	domainN := strings.Split(mdomain, ",")
	// fmt.Println(domainN)

	for _, d := range domainN {
		if strings.EqualFold(d, domain) {
			return true
		}
	}
	return false
}

func (this *SmtpdServer) cmdMailFrom(input string) bool {
	inputN := strings.SplitN(input, ":", 2)

	if len(inputN) == 2 {
		if this.cmdCompare(inputN[0], CMD_MAIL_FROM) {

			inputN[1] = strings.TrimSpace(inputN[1])
			inputN[1] = libs.FilterAddressBody(inputN[1])

			if !libs.CheckStandardMail(inputN[1]) {
				this.write(MSG_BAD_SYNTAX)
				return false
			}

			mailFrom := libs.GetRealMail(inputN[1])
			if !libs.IsEmailRe(mailFrom) {
				this.write(MSG_BAD_USER)
				return false
			}

			if this.isLogin && !this.runModeIn {
				info := strings.Split(mailFrom, "@")

				if !this.isAllowDomain(info[1]) {
					this.write(MSG_BAD_MAIL_ADDR)
					return false
				}

				user, err := db.UserGetByName(info[0])
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
	return false
}
func (this *SmtpdServer) cmdStartTtls(input string) bool {
	// if this.tls {
	// 	this.write(MSG_STARTTLS)
	// 	return true
	// }

	if this.TLSConfig == nil {
		this.w("502 Error: TLS not supported")
		return false
	}

	tlsConn := tls.Server(this.conn, this.TLSConfig)
	this.w("220 Go ahead\n")

	if err := tlsConn.Handshake(); err != nil {
		errmsg := fmt.Sprintf("550 ERROR: Handshake error:%s", err)
		this.w(errmsg)
		return false
	}

	state := tlsConn.ConnectionState()

	this.reader = bufio.NewReader(tlsConn)
	this.writer = bufio.NewWriter(tlsConn)
	this.scanner = bufio.NewScanner(this.reader)

	this.stateTLS = &state
	this.tls = true

	return true
}

func (this *SmtpdServer) cmdRcptTo(input string) bool {
	inputN := strings.SplitN(input, ":", 2)
	this.D("smtpd[cmd][rcpt to]", inputN[1])
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

			if this.runModeIn { //外部邮件,邮件地址检查
				info := strings.Split(rcptTo, "@")

				if !this.isAllowDomain(info[1]) {
					this.write(MSG_BAD_OPEN_RELAY)
					return false
				}
				user, err := db.UserGetByName(info[0])
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

func (this *SmtpdServer) addEnvelopeDataAcceptLine(data []byte) []byte {
	tlsDetails := ""

	tlsVersions := map[uint16]string{
		tls.VersionSSL30: "SSL3.0",
		tls.VersionTLS10: "TLS1.0",
		tls.VersionTLS11: "TLS1.1",
		tls.VersionTLS12: "TLS1.2",
		tls.VersionTLS13: "TLS1.3",
	}

	if this.stateTLS != nil {
		version := "unknown"

		if val, ok := tlsVersions[this.stateTLS.Version]; ok {
			version = val
		}

		cipher := tls.CipherSuiteName(this.stateTLS.CipherSuite)

		tlsDetails = fmt.Sprintf(
			"\r\n\t(version=%s cipher=%s);",
			version,
			cipher,
		)
	}

	peerIP := ""
	if addr, ok := this.peer.Addr.(*net.TCPAddr); ok {
		peerIP = addr.IP.String()
	}

	mdomain := config.GetString("mail.domain", "xxx.com")
	serverTagName := fmt.Sprintf("smtp.%s (NewMx)", mdomain)

	line := libs.Wrap([]byte(fmt.Sprintf(
		"Received: from %s (unknown[%s])\n\tby %s with SMTP id\n\tfor <%s>; %s %s\r\n",
		peerIP,
		peerIP,
		serverTagName,
		this.recordCmdMailFrom,
		time.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700 (MST)"),
		tlsDetails,
	)))

	data = append(data, line...)

	// Move the new Received line up front
	copy(data[len(line):], data[0:len(data)-len(line)])
	copy(data, line)
	return data

}

func (this *SmtpdServer) cmdDataAccept() bool {

	data := &bytes.Buffer{}
	reader := textproto.NewReader(this.reader).DotReader()
	_, err := io.CopyN(data, reader, int64(10240000))

	content := string(data.Bytes())
	this.D("smtpd[data]:", content)
	if err == io.EOF {
		this.write(MSG_MAIL_OK)
	}

	if this.runModeIn {
		fmt.Println("smtpd[data][peer]:", this.peer)
		revContent := string(this.addEnvelopeDataAcceptLine(data.Bytes()))
		_, err := db.MailPush(this.userID, 1, this.recordCmdMailFrom, this.recordcmdRcptTo, revContent, 3)
		if err != nil {
			return false
		}
	}

	if this.isLogin {

		_, err := db.MailPush(this.userID, 0, this.recordCmdMailFrom, this.recordcmdRcptTo, content, 0)
		if err != nil {
			return false
		}
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

// 本地用户邮件投递到其他邮件地址|需要登陆
func (this *SmtpdServer) cmdModeOut(state int, input string) bool {
	//CMD_AUTH_LOGIN
	if this.stateCompare(state, CMD_AUTH_LOGIN) {
		if this.cmdAuthLoginUser(input) {
			this.setState(CMD_AUTH_LOGIN_USER)
		}
	}

	//CMD_AUTH_LOGIN_USER
	if this.stateCompare(state, CMD_AUTH_LOGIN_USER) {
		if this.cmdQuit(input) {
			return true
		}

		if this.cmdAuthLoginPwd(input) {
			this.setState(CMD_AUTH_LOGIN_PWD)
		}
	}

	//CMD_AUTH_LOGIN_PWD
	if this.stateCompare(state, CMD_AUTH_LOGIN_PWD) {
		if this.cmdQuit(input) {
			return true
		}

		if this.cmdMailFrom(input) {
			this.setState(CMD_MAIL_FROM)
		}
	}

	return false
}

func (this *SmtpdServer) handle() {
	for {

		state := this.getState()

		input, err := this.getString(state)

		if err != nil {
			this.write(MSG_COMMAND_TM_CTC)
			this.close()
			break
		}

		this.D("smtpd[cmd]:", state, stateList[state], "input:[", input, "]")

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
				this.write(MSG_COMMAND_HE_ERR)
			}
		}

		//CMD_HELO
		if this.stateCompare(state, CMD_HELO) || this.stateCompare(state, CMD_EHLO) {

			if this.cmdQuit(input) {
				break
			} else if this.cmdHelo(input) {
				this.setState(CMD_HELO)
			} else if this.cmdEhlo(input) {
				this.setState(CMD_EHLO)
			}

			if this.modeIn {
				if this.cmdMailFrom(input) {
					this.setState(CMD_MAIL_FROM)
					this.runModeIn = true
				}
			} else {

				if this.cmdAuthPlainLogin(input) {
					this.setState(CMD_AUTH_LOGIN_PWD)
				} else if this.cmdAuthLogin(input) {
					this.setState(CMD_AUTH_LOGIN)
				}
			}
		}
		if this.runModeIn {
			this.D("当前运行模式：投递模式")
		} else {
			this.D("当前运行模式: 发送模式[需要认证]")
		}

		if this.enableStartTtls {
			if input == stateList[CMD_STARTTLS] { //CMD_STARTTLS
				if this.cmdStartTtls(input) {
					// this.write(MSG_STARTTLS)
				}
			}
		}

		if !this.runModeIn {
			isBreak := this.cmdModeOut(state, input)
			if isBreak {
				break
			}
		} else {
		} // 外部邮件投递到本地|不需要登陆

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

		if this.cmdQuit(input) {
			break
		}

		//CMD_DATA_END
		if this.stateCompare(state, CMD_DATA_END) {
			if this.cmdQuit(input) {
				break
			}
		}
	}
}

func (this *SmtpdServer) initTLSConfig() {
	this.TLSConfig = libs.InitAutoMakeTLSConfig()
}

func (this *SmtpdServer) ready() {
	this.initTLSConfig()

	this.startTime = time.Now()
	this.isLogin = false
	this.enableStartTtls = true

	//mode
	this.runModeIn = false
	this.modeIn, _ = config.GetBool("smtpd.mode_in", true)

}

func (this *SmtpdServer) start(conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(time.Minute * 30))
	this.conn = conn

	this.reader = bufio.NewReader(conn)
	this.writer = bufio.NewWriter(conn)
	this.scanner = bufio.NewScanner(this.reader)

	defer conn.Close()

	if this.enableStartTtls {
		var tlsConn *tls.Conn
		if tlsConn, this.tls = conn.(*tls.Conn); this.tls {
			tlsConn.Handshake()
			tlsState := tlsConn.ConnectionState()
			this.stateTLS = &tlsState
		}
	}

	this.peer = Peer{
		Addr: conn.RemoteAddr(),
		// ServerName: conn.Hostname,
	}

	this.write(MSG_INIT)
	this.setState(CMD_READY)

	this.handle()

}

func (this *SmtpdServer) StartPort(port int) {
	this.ready()

	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("StartPort", err)
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

func (this *SmtpdServer) StartSSLPort(port int) {
	this.ready()
	this.LinkSSL = true

	addr := fmt.Sprintf(":%d", port)
	ln, err := tls.Listen("tcp", addr, this.TLSConfig)
	if err != nil {
		fmt.Println("StartSSLPort", err)
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
	srv := SmtpdServer{}
	srv.StartPort(port)
}

func StartSSL(port int) {
	srvSSL := SmtpdServer{}
	srvSSL.StartSSLPort(port)
}
