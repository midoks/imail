package smtpd

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"

	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"github.com/midoks/imail/internal/config"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/libs"
	"io"
	"log"
	"net"
	"net/textproto"
	"strings"
	"time"
	// "strconv"
	// "errors"
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
	HeloName   string   // Server name used in HELO/EHLO command
	Username   string   // Username from authentication, if authenticated
	Password   string   // Password from authentication, if authenticated
	Protocol   Protocol // Protocol used, SMTP or ESMTP
	ServerName string   // A copy of Server.Hostname
	Addr       net.Addr // Network address
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

	reader  *bufio.Reader
	writer  *bufio.Writer
	scanner *bufio.Scanner

	//DB DATA
	userID int64

	//CMD_STARTTLS
	enableStartTtls bool

	//run mode
	modeIn bool

	// Determine the current mode of operation
	// 1,modeIn
	runModeIn bool

	peer Peer

	//tls
	AutoSSL   bool
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
	if this.AutoSSL {
		fmt.Print("[SSL]")
		return fmt.Println(a...)
	}
	return fmt.Println(a...)
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
		// fmt.Println("cmdAuthPlainLogin:", inputN)
		if len(inputN) == 3 {
			data := this.base64Decode(inputN[2])

			mdomain := config.GetString("mail.domain", "xxx.com")
			fmt.Println(mdomain)

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

func (this *SmtpdServer) cmdDataAccept() bool {

	data := &bytes.Buffer{}
	reader := textproto.NewReader(this.reader).DotReader()
	_, err := io.CopyN(data, reader, int64(10240000))

	content := string(data.Bytes())
	this.D("smtpd[data]", content)
	if err == io.EOF {
		this.write(MSG_MAIL_OK)
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
			} else if this.modeIn {
				if this.cmdMailFrom(input) {
					this.setState(CMD_MAIL_FROM)
					this.runModeIn = true
				}
			}

			if !this.runModeIn {

				if this.cmdAuthPlainLogin(input) {
					this.setState(CMD_AUTH_LOGIN_PWD)
				} else if this.cmdAuthLogin(input) {
					this.setState(CMD_AUTH_LOGIN)
				}
			}
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

	max := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, _ := rand.Int(rand.Reader, max)
	subject := pkix.Name{
		Organization:       []string{"IMAIL"},
		OrganizationalUnit: []string{"IMAIL ORG Unit"},
		CommonName:         "IMAIL",
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

	cert, err := tls.X509KeyPair(pem.EncodeToMemory(certBlock), pem.EncodeToMemory(privBlock))

	//------- demo start ------

	// 	var localhostCert = []byte(`-----BEGIN CERTIFICATE-----
	// MIIFkzCCA3ugAwIBAgIUQvhoyGmvPHq8q6BHrygu4dPp0CkwDQYJKoZIhvcNAQEL
	// BQAwWTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM
	// GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDESMBAGA1UEAwwJbG9jYWxob3N0MB4X
	// DTIwMDUyMTE2MzI1NVoXDTMwMDUxOTE2MzI1NVowWTELMAkGA1UEBhMCQVUxEzAR
	// BgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoMGEludGVybmV0IFdpZGdpdHMgUHR5
	// IEx0ZDESMBAGA1UEAwwJbG9jYWxob3N0MIICIjANBgkqhkiG9w0BAQEFAAOCAg8A
	// MIICCgKCAgEAk773plyfK4u2uIIZ6H7vEnTb5qJT6R/KCY9yniRvCFV+jCrISAs9
	// 0pgU+/P8iePnZRGbRCGGt1B+1/JAVLIYFZuawILHNs4yWKAwh0uNpR1Pec8v7vpq
	// NpdUzXKQKIqFynSkcLA8c2DOZwuhwVc8rZw50yY3r4i4Vxf0AARGXapnBfy6WerR
	// /6xT7y/OcK8+8aOirDQ9P6WlvZ0ynZKi5q2o1eEVypT2us9r+HsCYosKEEAnjzjJ
	// wP5rvredxUqb7OupIkgA4Nq80+4tqGGQfWetmoi3zXRhKpijKjgxBOYEqSUWm9ws
	// /aC91Iy5RawyTB0W064z75OgfuI5GwFUbyLD0YVN4DLSAI79GUfvc8NeLEXpQvYq
	// +f8P+O1Hbv2AQ28IdbyQrNefB+/WgjeTvXLploNlUihVhpmLpptqnauw/DY5Ix51
	// w60lHIZ6esNOmMQB+/z/IY5gpmuo66yH8aSCPSYBFxQebB7NMqYGOS9nXx62/Bn1
	// OUVXtdtrhfbbdQW6zMZjka0t8m83fnGw3ISyBK2NNnSzOgycu0ChsW6sk7lKyeWa
	// 85eJGsQWIhkOeF9v9GAIH/qsrgVpToVC9Krbk+/gqYIYF330tHQrzp6M6LiG5OY1
	// P7grUBovN2ZFt10B97HxWKa2f/8t9sfHZuKbfLSFbDsyI2JyNDh+Vk0CAwEAAaNT
	// MFEwHQYDVR0OBBYEFOLdIQUr3gDQF5YBor75mlnCdKngMB8GA1UdIwQYMBaAFOLd
	// IQUr3gDQF5YBor75mlnCdKngMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQEL
	// BQADggIBAGddhQMVMZ14TY7bU8CMuc9IrXUwxp59QfqpcXCA2pHc2VOWkylv2dH7
	// ta6KooPMKwJ61d+coYPK1zMUvNHHJCYVpVK0r+IGzs8mzg91JJpX2gV5moJqNXvd
	// Fy6heQJuAvzbb0Tfsv8KN7U8zg/ovpS7MbY+8mRJTQINn2pCzt2y2C7EftLK36x0
	// KeBWqyXofBJoMy03VfCRqQlWK7VPqxluAbkH+bzji1g/BTkoCKzOitAbjS5lT3sk
	// oCrF9N6AcjpFOH2ZZmTO4cZ6TSWfrb/9OWFXl0TNR9+x5c/bUEKoGeSMV1YT1SlK
	// TNFMUlq0sPRgaITotRdcptc045M6KF777QVbrYm/VH1T3pwPGYu2kUdYHcteyX9P
	// 8aRG4xsPGQ6DD7YjBFsif2fxlR3nQ+J/l/+eXHO4C+eRbxi15Z2NjwVjYpxZlUOq
	// HD96v516JkMJ63awbY+HkYdEUBKqR55tzcvNWnnfiboVmIecjAjoV4zStwDIti9u
	// 14IgdqqAbnx0ALbUWnvfFloLdCzPPQhgLHpTeRSEDPljJWX8rmy8iQtRb0FWYQ3z
	// A2wsUyutzK19nt4hjVrTX0At9ku3gMmViXFlbvyA1Y4TuhdUYqJauMBrWKl2ybDW
	// yhdKg/V3yTwgBUtb3QO4m1khNQjQLuPFVxULGEA38Y5dXSONsYnt
	// -----END CERTIFICATE-----`)

	// 	var localhostKey = []byte(`-----BEGIN PRIVATE KEY-----
	// MIIJQgIBADANBgkqhkiG9w0BAQEFAASCCSwwggkoAgEAAoICAQCTvvemXJ8ri7a4
	// ghnofu8SdNvmolPpH8oJj3KeJG8IVX6MKshICz3SmBT78/yJ4+dlEZtEIYa3UH7X
	// 8kBUshgVm5rAgsc2zjJYoDCHS42lHU95zy/u+mo2l1TNcpAoioXKdKRwsDxzYM5n
	// C6HBVzytnDnTJjeviLhXF/QABEZdqmcF/LpZ6tH/rFPvL85wrz7xo6KsND0/paW9
	// nTKdkqLmrajV4RXKlPa6z2v4ewJiiwoQQCePOMnA/mu+t53FSpvs66kiSADg2rzT
	// 7i2oYZB9Z62aiLfNdGEqmKMqODEE5gSpJRab3Cz9oL3UjLlFrDJMHRbTrjPvk6B+
	// 4jkbAVRvIsPRhU3gMtIAjv0ZR+9zw14sRelC9ir5/w/47Udu/YBDbwh1vJCs158H
	// 79aCN5O9cumWg2VSKFWGmYumm2qdq7D8NjkjHnXDrSUchnp6w06YxAH7/P8hjmCm
	// a6jrrIfxpII9JgEXFB5sHs0ypgY5L2dfHrb8GfU5RVe122uF9tt1BbrMxmORrS3y
	// bzd+cbDchLIErY02dLM6DJy7QKGxbqyTuUrJ5Zrzl4kaxBYiGQ54X2/0YAgf+qyu
	// BWlOhUL0qtuT7+CpghgXffS0dCvOnozouIbk5jU/uCtQGi83ZkW3XQH3sfFYprZ/
	// /y32x8dm4pt8tIVsOzIjYnI0OH5WTQIDAQABAoICADBPw788jje5CdivgjVKPHa2
	// i6mQ7wtN/8y8gWhA1aXN/wFqg+867c5NOJ9imvOj+GhOJ41RwTF0OuX2Kx8G1WVL
	// aoEEwoujRUdBqlyzUe/p87ELFMt6Svzq4yoDCiyXj0QyfAr1Ne8sepGrdgs4sXi7
	// mxT2bEMT2+Nuy7StsSyzqdiFWZJJfL2z5gZShZjHVTfCoFDbDCQh0F5+Zqyr5GS1
	// 6H13ip6hs0RGyzGHV7JNcM77i3QDx8U57JWCiS6YRQBl1vqEvPTJ0fEi8v8aWBsJ
	// qfTcO+4M3jEFlGUb1ruZU3DT1d7FUljlFO3JzlOACTpmUK6LSiRPC64x3yZ7etYV
	// QGStTdjdJ5+nE3CPR/ig27JLrwvrpR6LUKs4Dg13g/cQmhpq30a4UxV+y8cOgR6g
	// 13YFOtZto2xR+53aP6KMbWhmgMp21gqxS+b/5HoEfKCdRR1oLYTVdIxt4zuKlfQP
	// pTjyFDPA257VqYy+e+wB/0cFcPG4RaKONf9HShlWAulriS/QcoOlE/5xF74QnmTn
	// YAYNyfble/V2EZyd2doU7jJbhwWfWaXiCMOO8mJc+pGs4DsGsXvQmXlawyElNWes
	// wJfxsy4QOcMV54+R/wxB+5hxffUDxlRWUsqVN+p3/xc9fEuK+GzuH+BuI01YQsw/
	// laBzOTJthDbn6BCxdCeBAoIBAQDEO1hDM4ZZMYnErXWf/jik9EZFzOJFdz7g+eHm
	// YifFiKM09LYu4UNVY+Y1btHBLwhrDotpmHl/Zi3LYZQscWkrUbhXzPN6JIw98mZ/
	// tFzllI3Ioqf0HLrm1QpG2l7Xf8HT+d3atEOtgLQFYehjsFmmJtE1VsRWM1kySLlG
	// 11bQkXAlv7ZQ13BodQ5kNM3KLvkGPxCNtC9VQx3Em+t/eIZOe0Nb2fpYzY/lH1mF
	// rFhj6xf+LFdMseebOCQT27bzzlDrvWobQSQHqflFkMj86q/8I8RUAPcRz5s43YdO
	// Q+Dx2uJQtNBAEQVoS9v1HgBg6LieDt0ZytDETR5G3028dyaxAoIBAQDAvxEwfQu2
	// TxpeYQltHU/xRz3blpazgkXT6W4OT43rYI0tqdLxIFRSTnZap9cjzCszH10KjAg5
	// AQDd7wN6l0mGg0iyL0xjWX0cT38+wiz0RdgeHTxRk208qTyw6Xuh3KX2yryHLtf5
	// s3z5zkTJmj7XXOC2OVsiQcIFPhVXO3d38rm0xvzT5FZQH3a5rkpks1mqTZ4dyvim
	// p6vey4ZXdUnROiNzqtqbgSLbyS7vKj5/fXbkgKh8GJLNV4LMD6jo2FRN/LsEZKes
	// pxWNMsHBkv5eRfHNBVZuUMKFenN6ojV2GFG7bvLYD8Z9sja8AuBCaMr1CgHD8kd5
	// +A5+53Iva8hdAoIBAFU+BlBi8IiMaXFjfIY80/RsHJ6zqtNMQqdORWBj4S0A9wzJ
	// BN8Ggc51MAqkEkAeI0UGM29yicza4SfJQqmvtmTYAgE6CcZUXAuI4he1jOk6CAFR
	// Dy6O0G33u5gdwjdQyy0/DK21wvR6xTjVWDL952Oy1wyZnX5oneWnC70HTDIcC6CK
	// UDN78tudhdvnyEF8+DZLbPBxhmI+Xo8KwFlGTOmIyDD9Vq/+0/RPEv9rZ5Y4CNsj
	// /eRWH+sgjyOFPUtZo3NUe+RM/s7JenxKsdSUSlB4ZQ+sv6cgDSi9qspH2E6Xq9ot
	// QY2jFztAQNOQ7c8rKQ+YG1nZ7ahoa6+Tz1wAUnECggEAFVTP/TLJmgqVG37XwTiu
	// QUCmKug2k3VGbxZ1dKX/Sd5soXIbA06VpmpClPPgTnjpCwZckK9AtbZTtzwdgXK+
	// 02EyKW4soQ4lV33A0lxBB2O3cFXB+DE9tKnyKo4cfaRixbZYOQnJIzxnB2p5mGo2
	// rDT+NYyRdnAanePqDrZpGWBGhyhCkNzDZKimxhPw7cYflUZzyk5NSHxj/AtAOeuk
	// GMC7bbCp8u3Ows44IIXnVsq23sESZHF/xbP6qMTO574RTnQ66liNagEv1Gmaoea3
	// ug05nnwJvbm4XXdY0mijTAeS/BBiVeEhEYYoopQa556bX5UU7u+gU3JNgGPy8iaW
	// jQKCAQEAp16lci8FkF9rZXSf5/yOqAMhbBec1F/5X/NQ/gZNw9dDG0AEkBOJQpfX
	// dczmNzaMSt5wmZ+qIlu4nxRiMOaWh5LLntncQoxuAs+sCtZ9bK2c19Urg5WJ615R
	// d6OWtKINyuVosvlGzquht+ZnejJAgr1XsgF9cCxZonecwYQRlBvOjMRidCTpjzCu
	// 6SEEg/JyiauHq6wZjbz20fXkdD+P8PIV1ZnyUIakDgI7kY0AQHdKh4PSMvDoFpIw
	// TXU5YrNA8ao1B6CFdyjmLzoY2C9d9SDQTXMX8f8f3GUo9gZ0IzSIFVGFpsKBU0QM
	// hBgHM6A0WJC9MO3aAKRBcp48y6DXNA==
	// -----END PRIVATE KEY-----`)

	// 	cert, err := tls.X509KeyPair(localhostCert, localhostKey)

	//------- demo end ------

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

	this.reader = bufio.NewReader(conn)
	this.writer = bufio.NewWriter(conn)
	this.scanner = bufio.NewScanner(this.reader)

	defer conn.Close()

	this.peer = Peer{
		Addr: conn.RemoteAddr(),
		// ServerName: conn.Hostname,
	}

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

	if this.AutoSSL {
		this.cmdStartTtls("")
	}

	this.handle()

}

func (this *SmtpdServer) StartM(conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(time.Minute * 30))
	this.conn = conn

	this.reader = bufio.NewReader(conn)
	this.writer = bufio.NewWriter(conn)
	this.scanner = bufio.NewScanner(this.reader)

	defer conn.Close()

	this.peer = Peer{
		Addr: conn.RemoteAddr(),
		// ServerName: conn.Hostname,
	}

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

	if this.AutoSSL {
		this.cmdStartTtls("")
	}

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

func StartSSL(port int) {
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
		return
	}
	// ln.SetKeepAlivesEnabled(false)
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		srv_ssl := SmtpdServer{}
		srv_ssl.AutoSSL = true
		go srv_ssl.start(conn)
	}
}
