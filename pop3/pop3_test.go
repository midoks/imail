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
	// data, err2 = bufio.NewReader(conn).ReadString('\n')
	// if err2 != nil {
	// 	log.Println("ehlo directive error:", err2)
	// }
	// fmt.Println("S:", data)

	// data, err2 = bufio.NewReader(conn).ReadString('\n')
	// if err2 != nil {
	// 	log.Println("ehlo directive error:", err2)
	// }
	// fmt.Println("S:", data)

	// time.Sleep(time.Duration(2) * time.Second)

	//func 1
	// data2, err_2 := ioutil.ReadAll(conn)
	// chkError(err_2)
	// fmt.Println("ALL:", string(data2))

	//func 22
	// for {
	// 	data, err2 = bufio.NewReader(conn).ReadString('\n')
	// 	if err2 != nil {
	// 		log.Println("directive error:", err2)
	// 		break
	// 	}
	// 	fmt.Println("S--|:", data, err2)
	// }

	// data, err2 = bufio.NewReader(conn).ReadString('\n')
	// if err2 != nil {
	// 	log.Println("ehlo directive error:", err2)
	// }
	// fmt.Println("S:", data)

	// buf := make([]byte, 0, 4096)
	// var slen int

	// for {
	// 	n, err := conn.Read(buf)
	// 	// n, err := bufio.NewReader(conn).Read(buf)
	// 	fmt.Println(n, err)
	// 	if n > 0 {
	// 		slen += n
	// 	}

	// 	break
	// 	if err != nil {
	// 		if err == io.EOF {
	// 			break
	// 		}

	// 	}
	// }

	for {
		b := make([]byte, 4096)

		n, err2 := conn.Read(b[0:])
		fmt.Println(n, err)
		if err2 != nil {
			log.Println("directive error:", err2)
			break
		}
		fmt.Println("S--|~:", string(b[:n]))

		v := strings.TrimSpace(string(b[:n]))
		v_len := len(v)

		last_char := string(v[v_len-1 : v_len])

		fmt.Println("last char:", last_char)

		if strings.EqualFold(last_char, ".") {
			break
		}
	}

	conn.Write([]byte("QUIT\r\n"))
}

func TestRunPop3(t *testing.T) {
	PopList("pop3.163.com", "110")
}
