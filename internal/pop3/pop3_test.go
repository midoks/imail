package pop3

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/midoks/imail/internal/conf"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/log"
)

// go test -v ./internal/pop3
func init() {

	cDir, err := os.Getwd()
	appDir := filepath.Dir(filepath.Dir(cDir))

	os.Setenv("IMAIL_WORK_DIR", appDir)
	os.Chdir(appDir)
	err = conf.Init(appDir + "/custom/conf/app.conf")
	if err != nil {
		fmt.Println("TestReceivedMail config fail:", err.Error())
	}

	conf.Web.Domain = "cachecha.com"

	logger := log.Init()

	format := conf.Log.Format
	if strings.EqualFold(format, "json") {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else if strings.EqualFold(format, "text") {
		logger.SetFormatter(&logrus.TextFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{})
	}

	if strings.EqualFold(conf.App.RunMode, "dev") {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	go Start(10110)

	time.Sleep(1 * time.Second)
}

func PopCmd(domain string, port string, name string, password string) (bool, error) {
	var content string

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
	_, err = conn.Write([]byte(CMD_USER))
	if err != nil {
		return false, err
	}

	data, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return false, err
	}

	CMD_PWD := fmt.Sprintf("PASS %s\r\n", password)
	_, err = conn.Write([]byte(CMD_PWD))
	data, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return false, err
	}

	if strings.HasPrefix(data, "-ERR") {
		return false, errors.New(data)
	}

	_, err = conn.Write([]byte("STAT\r\n"))
	data, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return false, err
	}

	if strings.HasPrefix(data, "-ERR") {
		return false, errors.New(data)
	}

	page := ""
	cmd := fmt.Sprintf("LIST %s\r\n", page)
	// fmt.Println(cmd)
	_, err = conn.Write([]byte(cmd))

	if strings.EqualFold(page, "") {
		content = ""
		for {

			b := make([]byte, 4096)
			n, err := conn.Read(b[0:])
			if err != nil {
				break
			}

			v := strings.TrimSpace(string(b[:n]))
			content += fmt.Sprintf("%s\r\n", v)

			last := string(v[len(v)-1:])
			// fmt.Println("S-v:", v, "last:", last, "len:", len(v))
			if strings.EqualFold(last, ".") {
				break
			}
		}
		data = content
	} else {
		data, err = bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			return false, err
		}
	}

	_, err = conn.Write([]byte("UIDL 1\r\n"))
	data, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return false, err
	}

	if strings.HasPrefix(data, "-ERR") {
		return false, errors.New(data)
	}

	conn.Write([]byte("QUIT\r\n"))
	data, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return false, err
	}

	return true, nil
}

// go test -v pop3_test.go -test.run TestRunPop3
func D_TestRunPop3(t *testing.T) {
	// PopCmd("pop3.163.com", "110", "midoks", "mm123123")
}

// go test -v pop3_test.go -test.run TestRunLocalPop3
func TestRunLocalPop3(t *testing.T) {
	_, err := PopCmd("127.0.0.1", "10110", "admin", "admin")
	if err != nil {
		t.Error("TestRunLocalPop3 fail:" + err.Error())
	} else {
		t.Log("TestRunLocalPop3 ok")
	}
}
