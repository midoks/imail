package smtpd

import (
	"bufio"
	"fmt"
	// "log"
	"errors"
	"net"
	"strings"
)

const (
	deliveryDebug = true
)

func DeliveryDebug(args ...interface{}) {
	if deliveryDebug {
		fmt.Println("deliveryDebug:")
		fmt.Println(args...)
	}
}

// Delivery of mail to external mail
func Delivery(domain string, port string, from string, to string, content string) (bool, error) {

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

	data, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return false, err
	}

	if !strings.HasPrefix(data, "250") {
		return false, errors.New(data)
	}

	mailfrom := fmt.Sprintf("MAIL FROM: <%s>\r\n", from)
	DeliveryDebug(mailfrom)

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
	_, err = conn.Write([]byte(rcpt_to)) //向服务端发送数据。用n接受返回的数据大小，用err接受错误信息。
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

	DeliveryDebug(data)

	_, err = conn.Write([]byte("DATA\r\n")) //向服务端发送数据。用n接受返回的数据大小，用err接受错误信息。
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

	content = fmt.Sprintf("%s\r\n\r\n", content)
	DeliveryDebug(content)
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
