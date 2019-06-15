package imap

import (
	"bufio"
	"fmt"
	"net"
	"testing"
)

func imapCmd(domain string, port string, name string, password string) (bool, error) {
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
	fmt.Println("S:", data)

	CMD_USER := fmt.Sprintf("a1 %s %s\r\n", name, password)
	fmt.Println("C:", CMD_USER)
	_, err = conn.Write([]byte(CMD_USER))
	if err != nil {
		return false, err
	}

	return false, err
}

func TestRunImap(t *testing.T) {
	imapCmd("127.0.0.1", "143", "midoks", "mm123123")
}
