package task

import (
	"fmt"
	"github.com/midoks/imail/internal/config"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/libs"
	"github.com/midoks/imail/internal/smtpd"
	"github.com/robfig/cron"
	// "sync"
	"time"
)

func TaskQueueeSendMail() {
	domain := config.GetString("mail.domain", "xxx.com")
	postmaster := fmt.Sprintf("postmaster@%s", domain)

	content, err := libs.GetMailReturnToSender(postmaster, "admin@cachecha.com", "系统退信", "抱歉，您的邮件被退回来了……")
	fmt.Println(content, err)

	// err2 := smtpd.Delivery("127.0.0.1", postmaster, "admin@cachecha.com", []byte(content))
	// db.MailPush(1, 1, postmaster, "admin@cachecha.com", content, 0)
	// fmt.Println("err2", err2)
	result := db.MailSendListForStatus(2, 1)
	if len(result) > 0 {
		fmt.Println("email is doing!")
	} else {
		result = db.MailSendListForStatus(0, 1)
		for _, val := range result {
			db.MailSetStatusById(val.Id, 2)
			err := smtpd.Delivery("", val.MailFrom, val.MailTo, []byte(val.Content))
			if err == nil {
				db.MailSetStatusById(val.Id, 1)
			} else {

				err = smtpd.Delivery("", postmaster, val.MailTo, []byte("系统退信息"))
				fmt.Println("ddd:", err)
			}
			fmt.Println("send mail:", err)
		}
	}
}

func Init() {
	c := cron.New()

	c.AddFunc("*/5 * * * * * ", func() {
		fmt.Println(fmt.Sprintf("TaskQueueeSendMail! time:%d", time.Now().Unix()))
		TaskQueueeSendMail()
	})

	c.Start()
}
