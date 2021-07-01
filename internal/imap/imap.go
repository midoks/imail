package imap

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"

	"bufio"
	"fmt"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/imap/component"
	"github.com/midoks/imail/internal/libs"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	CMD_READY      = iota
	CMD_AUTH       = iota
	CMD_LIST       = iota
	CMD_XLIST      = iota
	CMD_LOGOUT     = iota
	CMD_CAPABILITY = iota
	CMD_ID         = iota
	CMD_STATUS     = iota
	CMD_SELECT     = iota
	CMD_FETCH      = iota
	CMD_UID        = iota
	CMD_COPY       = iota
	CMD_STORE      = iota
	CMD_NAMESPACE  = iota
	CMD_SEARCH     = iota
	CMD_NOOP       = iota
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
	CMD_STORE:      "STORE",
	CMD_NAMESPACE:  "NAMESPACE",
	CMD_SEARCH:     "SEARCH",
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

// var GO_EOL = "\n"

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

	TLSConfig *tls.Config // Enable STARTTLS support.
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

func (this *ImapServer) parseArgsConent(format string, data db.Mail) string {

	content := data.Content
	id := data.Id

	format = strings.TrimSpace(format)
	format = strings.Trim(format, "()")

	inputN := strings.Split(format, " ")
	list := make(map[string]interface{})

	bufferedBody := bufio.NewReader(strings.NewReader(content))
	header, err := component.ReadHeader(bufferedBody)

	if err != nil {
		fmt.Println("component.ReadHeader:", err)
	}

	bs, err := component.FetchBodyStructure(header, bufferedBody, true)

	// fmt.Println("FetchBodyStructure:", bs.ToString(), err)
	// fmt.Println("parseArgsConent[c][Mail]:", data)

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
			// fmt.Println("cmdUid[2]", inputN[2])
			// fmt.Println("cmdUid[3]", inputN[3])
			// fmt.Println("cmdUid[4]", inputN[4])
			if this.cmdCompare(inputN[2], CMD_FETCH) {

				if strings.Index(inputN[3], ":") > 0 {
					se := strings.SplitN(inputN[3], ":", 2)
					start, _ := strconv.ParseInt(se[0], 10, 64)
					end, _ := strconv.ParseInt(se[1], 10, 64)
					mailList, _ := db.BoxListBySE(this.userID, this.selectBox, start, end)
					for i, m := range mailList {
						c := this.parseArgsConent(inputN[4], m)
						this.writeArgs("* %d FETCH "+c, i+1)
					}
				}

				if libs.IsNumeric(inputN[3]) {
					mid, _ := strconv.ParseInt(inputN[3], 10, 64)
					mailList, _ := db.BoxListByMid(this.userID, this.selectBox, mid)
					c := this.parseArgsConent(inputN[4], mailList[0])
					this.writeArgs("* %d FETCH "+c, mid)
				}
			}

			if this.cmdCompare(inputN[2], CMD_SEARCH) {

				if strings.Index(inputN[4], ":") > 0 {
					se := strings.SplitN(inputN[4], ":", 2)
					start, _ := strconv.ParseInt(se[0], 10, 64)
					end, _ := strconv.ParseInt(se[1], 10, 64)
					mailList, _ := db.BoxListBySE(this.userID, this.selectBox, start, end)
					idString := ""
					for _, m := range mailList {
						idString += fmt.Sprintf(" %d", m.Id)
					}
					this.writeArgs("* SEARCH%s", idString)
				}

				if libs.IsNumeric(inputN[3]) {
					mid, _ := strconv.ParseInt(inputN[3], 10, 64)
					mailList, _ := db.BoxListByMid(this.userID, this.selectBox, mid)
					c := this.parseArgsConent(inputN[4], mailList[0])
					this.writeArgs("* %d SEARCH "+c, mid)
				}
			}

			if this.cmdCompare(inputN[2], CMD_COPY) {
				if libs.IsNumeric(inputN[3]) {
					mid, _ := strconv.ParseInt(inputN[3], 10, 64)
					inputN[4] = strings.Trim(inputN[4], "\"")
					if strings.EqualFold(inputN[4], "Deleted Messages") {
						db.MailSoftDeleteById(mid)
					}
				}
			}

			if this.cmdCompare(inputN[2], CMD_STORE) {
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

					if strings.EqualFold(inputN[5], "DELETED") && strings.HasPrefix(inputN[4], "+") {
						db.MailHardDeleteById(mid)
					}
				}
			}

			this.writeArgs("%s OK %s %s Completed", inputN[0], inputN[1], inputN[2])
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
		}

		if this.cmdId(input) {
		}

		if this.cmdNoop(input) {
		}

		if this.cmdAuth(input) {
			this.setState(CMD_AUTH)
		}

		if this.stateCompare(state, CMD_AUTH) {

			if this.cmdNameSpace(input) {

			}

			if this.cmdList(input) {

			}

			if this.cmdStatus(input) {

			}

			if this.cmdSelect(input) {

			}

			if this.cmdFecth(input) {

			}

			if this.cmdUid(input) {

			}

			if this.cmdLogout(input) {
				break
			}
		}

	}
}

func (this *ImapServer) initTLSConfig() {

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

func (this *ImapServer) StartSSLPort(port int) {
	this.initTLSConfig()
	addr := fmt.Sprintf(":%d", port)
	ln, err := tls.Listen("tcp", addr, this.TLSConfig)
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
		go this.start(conn)
	}
}

func Start(port int) {
	srv := ImapServer{}
	srv.StartPort(port)
}

func StartSSL(port int) {
	srv := ImapServer{}
	srv.StartSSLPort(port)
}
