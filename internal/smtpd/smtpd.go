package smtpd

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"github.com/midoks/imail/internal/config"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/libs"
	"log"
	"math/big"
	"net"
	"runtime"
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
	MSG_AUTH_LOGIN_USER = "334.user"
	MSG_AUTH_LOGIN_PWD  = "334.passwd"
	MSG_AUTH_OK         = "235"
	MSG_AUTH_FAIL       = "535"
	MSG_DATA            = "354"
	MSG_STARTTLS        = "220"
)

var msgList = map[string]string{
	MSG_INIT:            "Anti-spam GT for Coremail System(imail)",
	MSG_OK:              "ok",
	MSG_BYE:             "bye",
	MSG_COMMAND_HE_ERR:  "Error: send HELO/EHLO first",
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
	MSG_STARTTLS:        "Ready to start TLS from xxx to mail.xxx.com.",
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
	isLogin           bool
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

	//CMD_STARTTLS
	enableStartTtls bool

	//run mode
	modeIn bool

	// Determine the current mode of operation
	// 1,modeIn
	runModeIn bool

	//tls
	tls       bool
	stateTLS  *tls.ConnectionState
	TLSConfig *tls.Config // Enable STARTTLS support.
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
	fmt.Println("smtpd:", msg)
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

	// this.D("getString:", input, ":", err)
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
	if len(inputN) == 2 {
		if this.cmdCompare(inputN[0], CMD_HELO) {
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
			this.w("250-mail\r\n")
			this.w("250-PIPELINING\r\n")
			this.w("250-AUTH LOGIN PLAIN\r\n")
			this.w("250-AUTH=LOGIN\r\n")
			this.w("250-coremail 1Uxr2xKj7kG0xkI17xGrU7I0s8FY2U3Uj8Cz28x1UUUUU7Ic2I0Y2UFRbmXhUCa0xDrUUUUj\r\n")
			if this.enableStartTtls {
				this.w("250-STARTTLS\r\n")
			}

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

func (this *SmtpdServer) cmdAuthPlain(input string) bool {
	inputN := strings.SplitN(input, " ", 3)

	if len(inputN) == 3 {
		data := this.base64Decode(inputN[2])
		mdomain := config.GetString("mail.domain", "xxx.com")

		fmt.Println(mdomain)

		list := strings.SplitN(data, "@cachecha.com", 3)

		this.loginUser = list[0]
		this.loginPwd = list[2]

		b := this.checkUserLogin()
		this.D("smtpd:", b, this.loginUser, this.loginPwd)
		if b {
			this.write(MSG_AUTH_OK)
			return true
		}
		this.write(MSG_AUTH_FAIL)
	}
	return false
}

func (this *SmtpdServer) isAllowDomain(domain string) bool {
	mdomain := config.GetString("mail.domain", "xxx.com")
	domainN := strings.Split(mdomain, ",")
	fmt.Println(domainN)

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

			if !libs.CheckStandardMail(inputN[1]) {
				this.write(MSG_BAD_SYNTAX)
				return false
			}

			mailFrom := libs.GetRealMail(inputN[1])
			if !libs.IsEmailRe(mailFrom) {
				this.write(MSG_BAD_USER)
				return false
			}

			if this.isLogin {
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
	if this.tls {
		this.write(MSG_STARTTLS)
		return true
	}

	return false
}

func (this *SmtpdServer) cmdRcptTo(input string) bool {
	inputN := strings.SplitN(input, ":", 2)
	this.D("cmdRcptTo", inputN[1])
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

func (this *SmtpdServer) cmdDataAccept() bool {
	var content string
	content = ""
	for {

		b := make([]byte, 4096)
		n, _ := this.conn.Read(b[0:])

		line := strings.TrimSpace(string(b[:n]))
		content += fmt.Sprintf("%s\r\n", line)

		if line != "" {
			last := line[len(line)-1:]
			if strings.EqualFold(last, ".") {
				content = strings.TrimSpace(content[0 : len(content)-1])
				this.write(MSG_MAIL_OK)
				break
			}
		}
	}

	if this.runModeIn {
		_, err := db.MailPush(this.userID, 1, this.recordCmdMailFrom, this.recordcmdRcptTo, content, 3)
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

		if this.cmdAuthPlain(input) {
			this.setState(CMD_AUTH_LOGIN_PWD)
		} else if this.cmdAuthLoginPwd(input) {
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
			break
		}

		this.D("smtpd:", state, stateList[state], "input:[", input, "]")

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
			}

			if this.modeIn {
				if this.cmdMailFrom(input) {
					this.setState(CMD_MAIL_FROM)
					this.runModeIn = true
				}
			}

			if !this.runModeIn {
				if this.cmdAuthLogin(input) {
					this.setState(CMD_AUTH_LOGIN)
				}
			}
		}

		if this.enableStartTtls {
			if input == stateList[CMD_STARTTLS] { //CMD_STARTTLS

				if this.cmdStartTtls(input) {
					this.write(MSG_STARTTLS)
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

	max := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, _ := rand.Int(rand.Reader, max)
	subject := pkix.Name{
		Organization:       []string{"燕子李三"},
		OrganizationalUnit: []string{"Books"},
		CommonName:         "GO Web",
	}
	rootTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      subject,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	pk, _ := rsa.GenerateKey(rand.Reader, 2048)
	makeCert, _ := x509.CreateCertificate(rand.Reader, &rootTemplate, &rootTemplate, &pk.PublicKey, pk)

	privBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(pk),
	}

	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: makeCert,
	}

	// fmt.Println(string(pem.EncodeToMemory(privBlock)), string(pem.EncodeToMemory(certBlock)))
	cert, err := tls.X509KeyPair(pem.EncodeToMemory(certBlock), pem.EncodeToMemory(privBlock))
	if err != nil {
		log.Fatalf("Cert load failed: %v", err)
	}
	this.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

}
func (this *SmtpdServer) start(conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(time.Minute * 30))
	this.conn = conn
	defer conn.Close()
	this.startTime = time.Now()
	this.isLogin = false
	this.enableStartTtls = true

	if this.enableStartTtls {
		var tlsConn *tls.Conn
		if tlsConn, this.tls = conn.(*tls.Conn); this.tls {
			tlsConn.Handshake()
			tlsState := tlsConn.ConnectionState()
			this.stateTLS = &tlsState
		}
	}
	this.initTLSConfig()

	//mode
	this.runModeIn = false
	this.modeIn, _ = config.GetBool("smtpd.mode_in", false)

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
