package mail

import (
	"fmt"
	"github.com/midoks/imail/internal/conf"
	"github.com/midoks/imail/internal/tools"
	"io/ioutil"
	"regexp"
	"strings"
	"time"
)

func GetMailSubject(content string) string {
	var err error
	valid := regexp.MustCompile(`Subject: (.*)`)
	match := valid.FindAllStringSubmatch(content, -1)

	val := match[0][0]
	tmp := strings.Split(val, ":")
	val = strings.TrimSpace(tmp[1])

	if strings.Contains(val, "=?utf-8?B?") {
		val = strings.Replace(val, "=?utf-8?B?", "", -1)
		val = strings.Replace(val, "?=", "", -1)
		val = strings.TrimSpace(val)
		val, err = tools.Base64decode(val)
		// if err != nil {
		fmt.Println(val, err)
		// }
	}
	return val
}

func GetMailReturnToSender(to string, err_to_mail string, err_content string, msg string) (string, error) {
	sendSubject := GetMailSubject(err_content)

	domain := conf.Mail.Domain
	postmaster := fmt.Sprintf("postmaster@%s", domain)

	sendTime := time.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700 (MST)")
	sendVersion := fmt.Sprintf("imail/%s", conf.App.Version)
	boundaryRand := tools.RandString(20)

	data, err := ioutil.ReadFile(conf.WorkDir() + "conf/tpl/return_to_sender.tpl")
	if err != nil {
		return "", err
	}

	dataHtml, err := ioutil.ReadFile(conf.WorkDir() + "conf/tpl/return_to_sender_html.tpl")
	if err != nil {
		return "", err
	}

	contentHtml := strings.Replace(string(dataHtml), "{TILTE}", "邮箱退信", -1)
	contentHtml = strings.Replace(contentHtml, "{ERR_MSG}", msg, -1)
	contentHtml = strings.Replace(contentHtml, "{SEND_SUBJECT}", sendSubject, -1)
	contentHtml = strings.Replace(contentHtml, "{ERR_TO_MAIL}", err_to_mail, -1)

	content := strings.Replace(string(data), "{MAIL_FROM}", postmaster, -1)
	content = strings.Replace(content, "{RCPT_TO}", to, -1)
	content = strings.Replace(content, "{SUBJECT}", "系统退信", -1)
	content = strings.Replace(content, "{TIME}", sendTime, -1)
	content = strings.Replace(content, "{VERSION}", sendVersion, -1)
	content = strings.Replace(content, "{CONTENT}", tools.Base64encode(contentHtml), -1)
	content = strings.Replace(content, "{BOUNDARY_LINE}", boundaryRand, -1)
	return content, nil
}
