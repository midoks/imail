package libs

import (
	"fmt"
	"github.com/midoks/imail/internal/config"
	"io/ioutil"
	"strings"
	"time"
)

func GetMailReturnToSender(to string, err_to_mail string, msg string) (string, error) {

	domain := config.GetString("mail.domain", "xxx.com")
	postmaster := fmt.Sprintf("postmaster@%s", domain)

	sendTime := time.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700 (MST)")
	sendVersion := fmt.Sprintf("imail/%s", config.App.Version)
	boundaryRand := RandString(20)

	data, err := ioutil.ReadFile("conf/tpl/return_to_sender.tpl")
	if err != nil {
		return "", err
	}

	dataHtml, err := ioutil.ReadFile("conf/tpl/return_to_sender_html.tpl")
	if err != nil {
		return "", err
	}

	contentHtml := strings.Replace(string(dataHtml), "{TILTE}", "邮箱退信", -1)
	contentHtml = strings.Replace(contentHtml, "{ERR_MSG}", msg, -1)
	contentHtml = strings.Replace(contentHtml, "{ERR_TO_MAIL}", err_to_mail, -1)

	content := strings.Replace(string(data), "{MAIL_FROM}", postmaster, -1)
	content = strings.Replace(content, "{RCPT_TO}", to, -1)
	content = strings.Replace(content, "{SUBJECT}", "系统退信", -1)
	content = strings.Replace(content, "{TIME}", sendTime, -1)
	content = strings.Replace(content, "{VERSION}", sendVersion, -1)
	content = strings.Replace(content, "{CONTENT}", Base64encode(contentHtml), -1)
	content = strings.Replace(content, "{BOUNDARY_LINE}", boundaryRand, -1)
	return content, nil
}
