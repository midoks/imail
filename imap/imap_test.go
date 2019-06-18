package imap

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
)

func imapCmd(domain string, port string, name string, password string) (bool, error) {
	addr := fmt.Sprintf("%s:%s", domain, port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return false, err
	}
	defer conn.Close()
	var content string

	data, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return false, err
	}
	fmt.Println("S:", data)

	cmd := fmt.Sprintf("B CAPABILITY\r\n")
	fmt.Println("C:", cmd)
	_, err = conn.Write([]byte(cmd))
	if err != nil {
		return false, err
	}

	for {

		b := make([]byte, 4096)
		n, err := conn.Read(b[0:])
		if err != nil {
			break
		}

		v := strings.TrimSpace(string(b[:n]))
		content += fmt.Sprintf("%s\r\n", v)
		fmt.Println("S-v:", v)
		if strings.Contains(strings.ToLower(v), "completed") {
			break
		}
	}

	cmd = fmt.Sprintf("a1 login %s %s\r\n", name, password)
	fmt.Println("C:", cmd)
	_, err = conn.Write([]byte(cmd))
	if err != nil {
		return false, err
	}

	data, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return false, err
	}
	fmt.Println("S:", data)

	cmd = fmt.Sprintf("a1 LIST \"\" * \r\n")
	fmt.Println("C:", cmd)
	_, err = conn.Write([]byte(cmd))
	if err != nil {
		return false, err
	}

	for {

		b := make([]byte, 4096)
		n, err := conn.Read(b[0:])
		if err != nil {
			break
		}

		v := strings.TrimSpace(string(b[:n]))
		content += fmt.Sprintf("%s\r\n", v)
		fmt.Println("S-v:", v)

		if strings.Contains(strings.ToLower(v), "completed") {
			break
		}
	}

	cmd = fmt.Sprintf("a1 logout\r\n")
	fmt.Println("C:", cmd)
	_, err = conn.Write([]byte(cmd))
	if err != nil {
		return false, err
	}

	for {

		b := make([]byte, 4096)
		n, err := conn.Read(b[0:])
		fmt.Println(n, err)
		if err != nil {
			break
		}

		v := strings.TrimSpace(string(b[:n]))
		content += fmt.Sprintf("%s\r\n", v)
		fmt.Println("S-v:", v)
	}
	return false, err
}

// func TestRunImap(t *testing.T) {
// 	imapCmd("127.0.0.1", "143", "midoks", "mm123123")
// }

func TestRunImap163(t *testing.T) {
	imapCmd("imap.163.com", "143", "midoks@163.com", "mm123123")
}
