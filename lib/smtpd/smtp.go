package smtpd

import (
	// "errors"
	"fmt"
	// "math/rand"
	"net"
	// "strings"
	"bufio"
	// "io"
	"log"
)

func SendMail(domain string, from string, to string, content string) {

	addr := fmt.Sprintf("%s:25", domain)

	conn, err := net.Dial("tcp", addr) //拨号操作，需要指定协议。
	defer conn.Close()
	if err != nil {
		log.Fatal(err)
		return
	}

	conn.Write([]byte("EHLO 163.com\n")) //向服务端发送数据。用n接受返回的数据大小，用err接受错误信息。

	// if err != nil {
	// 	log.Fatal(err)
	// 	return
	// }

	data, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(data)

	mailfrom := fmt.Sprintf("MAIL FROM:<%s>\n", from)
	fmt.Println(mailfrom)

	_, err = conn.Write([]byte(mailfrom)) //向服务端发送数据。用n接受返回的数据大小，用err接受错误信息。
	if err != nil {
		log.Fatal(err)
		return
	}

	data2, err2 := bufio.NewReader(conn).ReadString('\n')
	if err2 != nil {
		log.Fatal(err2)
	}
	fmt.Println(data2)

	mailto := fmt.Sprintf("RCPT TO:<%s>\n", to)
	fmt.Println(to)

	_, err = conn.Write([]byte(mailto)) //向服务端发送数据。用n接受返回的数据大小，用err接受错误信息。
	if err != nil {
		log.Fatal(err)
		return
	}

	data3, err3 := bufio.NewReader(conn).ReadString('\n')
	if err3 != nil {
		log.Fatal(err3)
	}
	fmt.Println(data3)

	content = fmt.Sprintf("DATA\n%s", content)
	content = fmt.Sprintf("%s\n.\n", content)
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

	_, err = conn.Write([]byte("quit"))
	if err != nil {
		log.Fatal(err)
		return
	}

	data6, err6 := bufio.NewReader(conn).ReadString('\n')
	if err6 != nil {
		log.Fatal(err6)
	}
	fmt.Println(data6)

}
