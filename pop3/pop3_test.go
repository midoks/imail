package pop3

import (
	"bufio"
	"fmt"
	// "io"
	_ "io/ioutil"
	// "bytes"
	"log"
	"net"
	"strings"
	"testing"
	// "time"
)

func chkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func PopList(domain string, port string, name string, password string) {

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

	CMD_USER := fmt.Sprintf("USER %s\r\n", name)
	fmt.Println("CMD:", CMD_USER)
	_, err1 := conn.Write([]byte(CMD_USER))
	chkError(err1)
	data, err2 := bufio.NewReader(conn).ReadString('\n')
	if err2 != nil {
		log.Println("ehlo directive error:", err2)
	}

	fmt.Println("S:", data)

	CMD_PWD := fmt.Sprintf("PASS %s\r\n", password)
	fmt.Println("CMD:", CMD_PWD)
	_, err1 = conn.Write([]byte(CMD_PWD))
	chkError(err1)
	data, err2 = bufio.NewReader(conn).ReadString('\n')
	if err2 != nil {
		log.Println("ehlo directive error:", err2)
	}
	fmt.Println("S:", data)

	fmt.Println("CMD:LIST 1")
	_, err1 = conn.Write([]byte("LIST 1\r\n"))
	chkError(err1)
	data, err2 = bufio.NewReader(conn).ReadString('\n')
	if err2 != nil {
		log.Println("ehlo directive error:", err2)
	}
	fmt.Println("S:", data)

	fmt.Println("CMD:UIDL 1")
	_, err1 = conn.Write([]byte("UIDL 1\r\n"))
	chkError(err1)
	data, err2 = bufio.NewReader(conn).ReadString('\n')
	if err2 != nil {
		log.Println("ehlo directive error:", err2)
	}
	fmt.Println("S:", data)

	fmt.Println("CMD:RETR 1")
	_, err1 = conn.Write([]byte("RETR 1\r\n"))
	chkError(err1)

	for {

		b := make([]byte, 4096)

		n, err2 := conn.Read(b[0:])
		fmt.Println(n, err)
		if err2 != nil {
			log.Println("SS:directive error:", err2)
			break
		}
		// fmt.Println("S--|~:", string(b[:n]))

		v := strings.TrimSpace(string(b[:n]))
		v_len := len(v)

		last_char := string(v[v_len-1 : v_len])

		fmt.Println("last char:", last_char)

		if strings.EqualFold(last_char, ".") {
			break
		}
	}

	conn.Write([]byte("QUIT\r\n"))

	data, err2 = bufio.NewReader(conn).ReadString('\n')
	if err2 != nil {
		log.Println("ehlo directive error:", err2)
	}
	fmt.Println("S:", data)
}

func TestRunPop3(t *testing.T) {
	PopList("pop3.163.com", "110", "midoks", "mm123123")
}

func TestRunLocalPop3(t *testing.T) {
	fmt.Println("---------------------------------")
	PopList("127.0.0.1", "10110", "midoks", "123123")
}
