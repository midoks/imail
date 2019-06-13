package smtpd

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/midoks/imail/libs"
	"net"
	"strings"
	"testing"
)

// Delivery of mail to external mail
func SendMail(user, pwd, domain string, port string, from string, to string, subject string, msg string) (bool, error) {

	addr := fmt.Sprintf("%s:%s", domain, port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	data, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return false, err
	}

	_, err = conn.Write([]byte("EHLO IMAIL\r\n"))
	if err != nil {
		return false, err
	}

	data = ""
	for {

		b := make([]byte, 4096)
		n, err := conn.Read(b[0:])
		fmt.Println(n, err)
		if err != nil {
			break
		}

		v := strings.TrimSpace(string(b[:n]))
		data += fmt.Sprintf("%s\r\n", v)
		fmt.Println(v)
		// last := string(v[0:4])
		// fmt.Println(last)
		inputN := strings.Split(v, "\r\n")

		// for i := 0; i < len(inputN); i++ {
		// 	fmt.Println("dd:v", inputN[i])
		// }

		// fmt.Println(inputN, len(inputN))
		last := inputN[len(inputN)-1:][0]
		fmt.Println(last)
		if strings.EqualFold(last, "250 8BITMIME") {
			break
		}
	}

	fmt.Println(data)

	// if !strings.HasPrefix(data, "250") {
	// 	return false, errors.New(data)
	// }

	_, err = conn.Write([]byte("AUTH LOGIN\r\n"))
	if err != nil {
		return false, err
	}

	data, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return false, err
	}
	fmt.Println(data)

	if !strings.HasPrefix(data, "334") {
		return false, errors.New(data)
	}

	user_input := fmt.Sprintf("%s\r\n", libs.Base64encode(user))
	_, err = conn.Write([]byte(user_input))
	if err != nil {
		return false, err
	}

	data, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return false, err
	}
	fmt.Println(data)
	if !strings.HasPrefix(data, "334") {
		return false, errors.New(data)
	}

	pwd_input := fmt.Sprintf("%s\r\n", libs.Base64encode(pwd))
	fmt.Println(pwd_input)
	_, err = conn.Write([]byte(pwd_input))
	if err != nil {
		return false, err
	}

	data, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return false, err
	}

	if !strings.HasPrefix(data, "235") {
		return false, errors.New(data)
	}

	mailfrom := fmt.Sprintf("MAIL FROM: <%s>\r\n", from)
	_, err = conn.Write([]byte(mailfrom))
	if err != nil {
		return false, err
	}

	data, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return false, err
	}

	if !strings.HasPrefix(data, "250") {
		return false, errors.New(data)
	}

	rcpt_to := fmt.Sprintf("RCPT TO: <%s>\r\n", to)
	DeliveryDebug(rcpt_to)
	_, err = conn.Write([]byte(rcpt_to))
	if err != nil {
		return false, err
	}

	data, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return false, err
	}

	if !strings.HasPrefix(data, "250") {
		return false, errors.New(data)
	}

	_, err = conn.Write([]byte("DATA\r\n"))
	if err != nil {
		return false, err
	}

	data, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return false, err
	}
	if !strings.HasPrefix(data, "354") {
		return false, errors.New(data)
	}

	content := fmt.Sprintf("From: <%s>\r\n", from)
	content += fmt.Sprintf("To: <%s>\r\n", to)
	content += fmt.Sprintf("Subject: %s\r\n\r\n", subject)
	content += fmt.Sprintf("%s\r\n", msg)
	_, err = conn.Write([]byte(content))
	if err != nil {
		return false, err
	}

	_, err = conn.Write([]byte(".\r\n"))
	if err != nil {
		return false, err
	}

	data, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return false, err
	}
	if !strings.HasPrefix(data, "354") {
		return false, errors.New(data)
	}

	_, err = conn.Write([]byte("QUIT\r\n"))
	if err != nil {
		return false, err
	}

	data, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return false, err
	}

	if !strings.HasPrefix(data, "221") {
		return false, errors.New(data)
	}

	return true, nil
}

// func TestDnsQuery(t *testing.T) {
// 	d, err := DnsQuery("163.com")
// 	fmt.Println(d, err)
// 	if err == nil {
// 		t.Log("dns.Query ok:" + d)
// 	} else {
// 		t.Error("dns.Query fail:" + err.Error())
// 	}
// }

func mailDeliveryTest() {
	toEmail := "midoks@163.com"
	fromEmail := "midoks@cachecha.com"
	toInfo := strings.Split(toEmail, "@")
	mxDomain, err := DnsQuery(toInfo[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(mxDomain)

	content := fmt.Sprintf("From: <%s>\r\nSubject: Hello imail\r\nTo: <%s>\r\n\r\nHi! yes is test. imail ok?", fromEmail, toEmail)
	_, err = Delivery(mxDomain, "25", fromEmail, toEmail, content)
	if err != nil {
		fmt.Println("err:", err)
	}
}

func mailDeliveryTest2() {
	toEmail := "midoks@1632.com"
	fromEmail := "midoks@cachecha.com"
	mxDomain, err := DnsQuery("163.com")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(mxDomain)

	content := fmt.Sprintf("From: <%s>\r\nSubject: Hello imail\r\nTo: <%s>\r\n\r\nHi! yes is test. imail ok?", fromEmail, toEmail)
	_, err = Delivery(mxDomain, "25", fromEmail, toEmail, content)
	if err != nil {
		fmt.Println("err:", err)
	}
}

func TestRunSendDelivery(t *testing.T) {
	// sendMailTest2()
}

func TestRunUserSend(t *testing.T) {
	_, err := SendMail("midoks", "123123", "127.0.0.1", "1025", "midoks@imail.com", "midoks@163.com", "title test!", "content is test!")
	if err != nil {
		fmt.Println("err:", err)
	}

	// _, err = SendMail("midoks", "mm123123", "smtp.163.com", "25", "midoks@163.com", "627293072@qq.com", "php求增加pcntl扩展!", "谢谢使用，我有空了就加上吧!")
	// if err != nil {
	// 	fmt.Println("err:", err)
	// }
}

func TestRunSendLocal(t *testing.T) {
	toEmail := "midoks@cachecha.com"
	fromEmail := "midoks@11.com"
	content := fmt.Sprintf("From: <%s>\r\nSubject: Hello imail\r\nTo: <%s>\r\n\r\nHi! yes is test. imail ok?!", fromEmail, toEmail)
	_, err := Delivery("127.0.0.1", "1025", fromEmail, toEmail, content)
	if err != nil {
		t.Error(err)
	}
}

// func TestRunSendFuncQQ(t *testing.T) {
// 	toEmail := "627293072@qq.com"
// 	fromEmail := "midoks@163.com"
// 	toInfo := strings.Split(toEmail, "@")
// 	mxDomain, err := DnsQuery(toInfo[1])
// 	if err != nil {
// 		fmt.Println(err.Error())
// 		return
// 	}
// 	fmt.Println(mxDomain)

// 	content := fmt.Sprintf("From: <121212312@qq.com>\r\nSubject: Hello imail\r\nTo: <%s>\r\n\r\nHi! yes is test. imail ok?!", toEmail)
// 	_, err = Delivery(mxDomain, "25", fromEmail, toEmail, content)
// 	if err != nil {
// 		fmt.Println("err:", err)
// 	}

// 	fmt.Println("-------------qq----end----------------")
// }

func Benchmark_SendLocal(b *testing.B) {
	toEmail := "midoks@imail.com"
	fromEmail := "midoks@cachecha.com"
	content := fmt.Sprintf("From: <%s>\r\nSubject: Hello imail\r\nTo: <%s>\r\n\r\nHi! yes is test. imail ok?!", fromEmail, toEmail)
	for i := 0; i < b.N; i++ { //use b.N for looping
		Delivery("127.0.0.1", "1025", fromEmail, toEmail, content)
	}
}
