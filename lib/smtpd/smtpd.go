package smtpd

import (
	// "errors"
	"fmt"
	// "math/rand"
	"bufio"
	"net"
	"strings"
	"time"
	// "io"
	"log"
)

const (
	HELO      = 1
	MAIL_FROM = 2
	RCPT_TO   = 3
	DATA      = 4
	QUIT      = 5
)

var stateList = map[int]string{
	HELO:      "HELO",
	MAIL_FROM: "MAIL_FROM",
	RCPT_TO:   "RCPT_TO",
	DATA:      "DATA",
	QUIT:      "QUIT",
}

const (
	INIT        = 222
	OK          = 220
	BYE         = 221
	COMMAND_ERR = 502
)

var msgList = map[int]string{
	INIT:        "Anti-spam GT for Coremail System(imail)",
	OK:          "ok",
	BYE:         "bye",
	COMMAND_ERR: "Error: command not implemented",
}

type smtpdServer struct {
	state     int
	inputTime time.Time
}

func (c *smtpdServer) setState(state int) {
	c.state = state
}

func (c *smtpdServer) D(a ...interface{}) (n int, err error) {
	return fmt.Println(a...)
}

func (c *smtpdServer) write(conn net.Conn, code int) {

	info := fmt.Sprintf("%d %s\n", code, msgList[code])
	_, err := conn.Write([]byte(info))

	if err != nil {
		log.Fatal(err)
	}
}

func (c *smtpdServer) writeString(conn net.Conn, code int, msg string) {

	info := fmt.Sprintf("%d %s\n", code, msg)
	_, err := conn.Write([]byte(info))

	if err != nil {
		log.Fatal(err)
	}
}

func (c *smtpdServer) guess(conn net.Conn, state int) bool {

	input, err := bufio.NewReader(conn).ReadString('\n')
	input = strings.Trim(input, "\n")

	if err != nil {
		log.Fatal(err)
		return false
	}

	c.D(input, stateList[state], strings.EqualFold(input, stateList[state]))

	if strings.EqualFold(input, stateList[state]) {
		c.setState(state)
		return true
	}
	return false
}

func (c *smtpdServer) getString(conn net.Conn) (string, error) {

	input, err := bufio.NewReader(conn).ReadString('\n')

	if err != nil {
		return "", err
	}

	input = strings.Trim(input, "\n")
	fmt.Println(input)

	return input, err
}

func (c *smtpdServer) handle(conn net.Conn) {
	c.write(conn, INIT)
	for {

	GOHELO:
		if !c.guess(conn, HELO) {
			c.write(conn, COMMAND_ERR)
			goto GOHELO
		}
		c.write(conn, OK)

		if c.guess(conn, QUIT) {
			defer conn.Close()
			break
		}
	}
}

func (c *smtpdServer) start() {

	ln, err := net.Listen("tcp", ":1025")
	if err != nil {
		panic(err)
		return
	}

	for {
		conn, err := ln.Accept()

		if err != nil {
			continue
		}
		go c.handle(conn)
	}
}

func Start() {
	s := smtpdServer{}
	s.start()
	fmt.Println("start!!!")
}
