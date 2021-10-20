package mail

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/midoks/imail/internal/conf"
	"github.com/midoks/imail/internal/tools"
)

func GetMailSubject(content string) string {
	var err error
	valid := regexp.MustCompile(`Subject: (.*)`)
	match := valid.FindAllStringSubmatch(content, -1)

	val := match[0][0]
	tmp := strings.SplitN(val, ":", 2)
	val = strings.TrimSpace(tmp[1])

	if strings.Contains(val, "=?utf-8?B?") || strings.Contains(val, "=?UTF-8?B?") {
		val = strings.Replace(val, "=?utf-8?B?", "", -1)
		val = strings.Replace(val, "=?UTF-8?B?", "", -1)
		val = strings.Replace(val, "?=", "", -1)
		val = strings.TrimSpace(val)
		val, err = tools.Base64decode(val)
		if err == nil {
			return val
		}
	}

	if strings.Contains(val, "=?gbk?B?") || strings.Contains(val, "=?GBK?B?") {
		val = strings.Replace(val, "=?gbk?B?", "", -1)
		val = strings.Replace(val, "=?GBK?B?", "", -1)
		val = strings.Replace(val, "?=", "", -1)
		val = strings.TrimSpace(val)
		val, err = tools.Base64decode(val)
		if err == nil {
			val = tools.ConvertToString(val, "gbk", "utf-8")
			return val
		}
	}
	return val
}

func GetMailFromInContent(content string) string {
	var err error
	valid := regexp.MustCompile(`From: (.*)`)
	match := valid.FindAllStringSubmatch(content, -1)

	val := match[0][0]
	tmp := strings.SplitN(val, ":", 2)
	val = strings.TrimSpace(tmp[1])

	tmp = strings.SplitN(val, "<", 2)
	val = strings.TrimSpace(tmp[0])
	val = strings.Trim(val, "\"")

	if strings.EqualFold(val, "") {
		val = tmp[1]
		val = strings.Trim(val, ">")
		tmp = strings.SplitN(val, "@", 2)
		return tmp[0]
	}

	if strings.Contains(val, "=?utf-8?B?") || strings.Contains(val, "=?UTF-8?B?") {
		val = strings.Replace(val, "=?utf-8?B?", "", -1)
		val = strings.Replace(val, "=?UTF-8?B?", "", -1)
		val = strings.Replace(val, "?=", "", -1)
		val = strings.TrimSpace(val)
		val, err = tools.Base64decode(val)
		if err == nil {
			return val
		}
	}
	return val
}

func GetMailRealContent(from, to, subject, content string) string {
	tplContent := conf.MustAsset("conf/tpl/send.tpl")
	sendTime := time.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700 (MST)")
	sendVersion := fmt.Sprintf("imail/%s", conf.App.Version)
	boundaryRand := tools.RandString(20)

	tplContent = bytes.Replace(tplContent, []byte("{VERSION}"), []byte(sendVersion), -1)
	tplContent = bytes.Replace(tplContent, []byte("{TIME}"), []byte(sendTime), -1)
	tplContent = bytes.Replace(tplContent, []byte("{BOUNDARY_LINE}"), []byte(boundaryRand), -1)
	tplContent = bytes.Replace(tplContent, []byte("{MAIL_FROM}"), []byte(from), -1)
	tplContent = bytes.Replace(tplContent, []byte("{RCPT_TO}"), []byte(to), -1)
	tplContent = bytes.Replace(tplContent, []byte("{SUBJECT}"), []byte(subject), -1)
	tplContent = bytes.Replace(tplContent, []byte("{CONTENT}"), []byte(tools.Base64encode(content)), -1)
	return string(tplContent)
}

func GetMailReturnToSender(to string, err_to_mail string, err_content string, msg string) (string, error) {
	sendSubject := GetMailSubject(err_content)

	domain := conf.Web.Domain
	postmaster := fmt.Sprintf("postmaster@%s", domain)

	sendTime := time.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700 (MST)")
	sendVersion := fmt.Sprintf("imail/%s", conf.App.Version)
	boundaryRand := tools.RandString(20)
	tmp := conf.MustAsset("conf/tpl/return_to_sender.tpl")
	data := string(tmp)

	tmp2 := conf.MustAsset("conf/tpl/return_to_sender_html.tpl")
	dataHtml := string(tmp2)

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
