package pop3

import (
	"bufio"
	"fmt"
	// "io"
	// "io/ioutil"
	"log"
	"net"
	// "strings"
	"testing"
)

func chkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func PopList(domain string, port string) {

	addr := fmt.Sprintf("%s:%s", domain, port)
	conn, err := net.Dial("tcp", addr) //拨号操作，需要指定协议。

	if err != nil {
		log.Println("dial error:", err)
		return
	}
	defer conn.Close()

	connect, err0 := bufio.NewReader(conn).ReadString('\n')
	if err0 != nil {
		log.Println("telnet:", err0)
	}
	fmt.Println("S:", connect)

	_, err1 := conn.Write([]byte("USER midoks\r\n"))
	chkError(err1)
	data, err2 := bufio.NewReader(conn).ReadString('\n')
	if err2 != nil {
		log.Println("ehlo directive error:", err2)
	}

	fmt.Println("S:", data)

	_, err1 = conn.Write([]byte("PASS mm123123\r\n"))
	chkError(err1)
	data, err2 = bufio.NewReader(conn).ReadString('\n')
	if err2 != nil {
		log.Println("ehlo directive error:", err2)
	}
	fmt.Println("S:", data)

	_, err1 = conn.Write([]byte("LIST 1\r\n"))
	chkError(err1)
	data, err2 = bufio.NewReader(conn).ReadString('\n')
	if err2 != nil {
		log.Println("ehlo directive error:", err2)
	}
	fmt.Println("S:", data)

	_, err1 = conn.Write([]byte("UIDL 1\r\n"))
	chkError(err1)
	data, err2 = bufio.NewReader(conn).ReadString('\n')
	if err2 != nil {
		log.Println("ehlo directive error:", err2)
	}
	fmt.Println("S:", data)

	_, err1 = conn.Write([]byte("RETR 1\r\n"))
	chkError(err1)
	data, err2 = bufio.NewReader(conn).ReadString('\n')
	if err2 != nil {
		log.Println("ehlo directive error:", err2)
	}
	fmt.Println("S:", data)

	_, err1 = conn.Write([]byte("Receive\r\n"))
	chkError(err1)

	// buf := make([]byte, 0, 4096)
	// buf, err = ioutil.ReadAll(conn)
	// fmt.Println(buf, err)

	data, err2 = bufio.NewReader(conn).ReadString('\n')
	if err2 != nil {
		log.Println("ehlo directive error:", err2)
	}
	fmt.Println("S:", data)

	conn.Write([]byte("QUIT\r\n"))
}

func TestRunPop3(t *testing.T) {
	PopList("pop3.163.com", "110")
}
