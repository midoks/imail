package libs

import (
	"fmt"
	"github.com/midoks/imail/internal/config"
	"io/ioutil"
	"strings"
	"time"
)

func GetMailReturnToSender(from string, to string, subject string, msg string) (string, error) {
	data, err := ioutil.ReadFile("conf/tpl/return_to_sender.tpl")
	if err != nil {
		return "", err
	}

	sendTime := time.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700 (MST)")
	sendVersion := fmt.Sprintf("imail/%s", config.App.Version)

	content := strings.Replace(string(data), "{MAIL_FROM}", from, -1)
	content = strings.Replace(content, "{RCPT_TO}", to, -1)
	content = strings.Replace(content, "{SUBJECT}", subject, -1)
	content = strings.Replace(content, "{TIME}", sendTime, -1)
	content = strings.Replace(content, "{VERSION}", sendVersion, -1)
	content = strings.Replace(content, "{CONTENT}", msg, -1)
	return content, nil
}
