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

func PopCmd(domain string, port string, name string, password string) (bool, error) {

	addr := fmt.Sprintf("%s:%s", domain, port)
	conn, err := net.Dial("tcp", addr) //拨号操作，需要指定协议。
	if err != nil {
		return false, err
	}
	defer conn.Close()

	data, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return false, err
	}

	CMD_USER := fmt.Sprintf("USER %s\r\n", name)
	fmt.Println("CMD:", CMD_USER)
	_, err = conn.Write([]byte(CMD_USER))
	if err != nil {
		return false, err
	}

	data, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return false, err
	}

	fmt.Println("S:", data)

	CMD_PWD := fmt.Sprintf("PASS %s\r\n", password)
	fmt.Println("CMD:", CMD_PWD)
	_, err = conn.Write([]byte(CMD_PWD))
	data, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return false, err
	}
	fmt.Println("S:", data)

	fmt.Println("CMD:LIST 1")

	_, err = conn.Write([]byte("LIST 1\r\n"))
	data, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return false, err
	}
	fmt.Println("S:", data)

	fmt.Println("CMD:UIDL 1")
	_, err = conn.Write([]byte("UIDL 1\r\n"))
	data, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return false, err
	}
	fmt.Println("S:", data)

	fmt.Println("CMD:RETR 1")
	_, err = conn.Write([]byte("RETR 1\r\n"))

	for {

		b := make([]byte, 4096)

		n, err := conn.Read(b[0:])
		fmt.Println(n, err)
		if err != nil {
			log.Println("SS:directive error:", err)
			break
		}

		v := strings.TrimSpace(string(b[:n]))

		fmt.Println("line:", v)
		last := string(v[len(v)-1:])

		if strings.EqualFold(last, ".") {
			break
		}
	}

	conn.Write([]byte("QUIT\r\n"))

	data, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return false, err
	}
	fmt.Println("S:", data)

	return true, nil
}

// go test -v pop3_test.go -test.run TestRunPop3
func TestRunPop3(t *testing.T) {
	PopCmd("pop3.163.com", "110", "midoks", "mm123123")
}

// go test -v pop3_test.go -test.run TestRunLocalPop3
func TestRunLocalPop3(t *testing.T) {

	_, err := PopCmd("127.0.0.1", "10110", "midoks", "123123")
	if err != nil {
		fmt.Println(err)
	}
}
