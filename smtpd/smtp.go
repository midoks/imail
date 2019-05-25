package smtpd

import (
	// "errors"
	"fmt"
	// "math/rand"
	"bufio"
	"net"
	"strings"
	// "io"
	"log"
)

func chkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func SendMail(domain string, from string, to string, content string) {

	addr := fmt.Sprintf("%s:25", domain)

	conn, err := net.Dial("tcp", addr) //拨号操作，需要指定协议。

	if err != nil {
		log.Println("dial error:", err)
		return
	}
	defer conn.Close()

	connect, err0 := bufio.NewReader(conn).ReadString('\n')
	if err0 != nil {
		log.Println("connect directive error:", err0)
	}
	fmt.Println("connect info:", connect)

	_, err1 := conn.Write([]byte("EHLO IMAIL\r\n"))
	chkError(err1)

	data, err2 := bufio.NewReader(conn).ReadString('\n')
	if err2 != nil {
		log.Println("ehlo directive error:", err2)
	}
	fmt.Println("ehlo info", data)

	mailfrom := fmt.Sprintf("MAIL FROM:<%s>\r\n", from)
	fmt.Println(mailfrom)

	conn.Write([]byte(mailfrom))
	data2, err3 := bufio.NewReader(conn).ReadString('\n')
	if err2 != nil {
		log.Println("mail from directive error:", err3)
		return
	}
	if !strings.HasPrefix(data2, "250") {
		fmt.Println("ddd:", data2)
		return
	}

	rcpt_to := fmt.Sprintf("RCPT TO:<%s>\r\n", to)
	fmt.Println(rcpt_to)

	rcpt_to_data, err := conn.Write([]byte(rcpt_to)) //向服务端发送数据。用n接受返回的数据大小，用err接受错误信息。
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println(rcpt_to_data)

	data3, err3 := bufio.NewReader(conn).ReadString('\n')
	if err3 != nil {
		log.Fatal(err3)
		return
	}
	fmt.Println(data3)

	content = fmt.Sprintf("DATA\r\n%s", content)
	content = fmt.Sprintf("%s\r\n.\r\n", content)
	fmt.Println(content)

	_, err = conn.Write([]byte(content)) //向服务端发送数据。用n接受返回的数据大小，用err接受错误信息。
	if err != nil {
		log.Fatal(err)
		return
	}

	// data4, err4 := bufio.NewReader(conn).ReadString('\n')
	// if err4 != nil {
	// 	log.Fatal(err4)
	// }
	// fmt.Println(data4)

	// _, err = conn.Write([]byte(content)) //向服务端发送数据。用n接受返回的数据大小，用err接受错误信息。
	// if err != nil {
	// 	log.Fatal(err)
	// 	return
	// }

	data5, err5 := bufio.NewReader(conn).ReadString('\n')
	if err5 != nil {
		log.Fatal(err5)
	}
	fmt.Println(data5)

	_, err = conn.Write([]byte("quit\r\n"))
	if err != nil {
		log.Fatal(err)
		return
	}

	data6, err6 := bufio.NewReader(conn).ReadString('\n')
	if err6 != nil {
		log.Fatal(err6)
		return
	}
	fmt.Println(data6)

}
