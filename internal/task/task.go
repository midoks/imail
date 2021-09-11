package task

import (
	"fmt"
	"github.com/midoks/imail/internal/config"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/libs"
	"github.com/midoks/imail/internal/log"
	"github.com/midoks/imail/internal/smtpd"
	"github.com/robfig/cron"
	// "sync"
	// "os"
)

func TaskQueueeSendMail() {
	domain := config.GetString("mail.domain", "xxx.com")
	postmaster := fmt.Sprintf("postmaster@%s", domain)

	result := db.MailSendListForStatus(2, 1)
	if len(result) == 0 {

		result = db.MailSendListForStatus(0, 1)
		for _, val := range result {
			db.MailSetStatusById(val.Id, 2)
			err := smtpd.Delivery("", val.MailFrom, val.MailTo, []byte(val.Content))
			if err != nil {

				content, _ := libs.GetMailReturnToSender(val.MailFrom, val.MailTo, val.Content, err.Error())
				db.MailPush(val.Uid, 1, postmaster, val.MailFrom, content, 1)
			}
			db.MailSetStatusById(val.Id, 1)
			// fmt.Println("send mail:", err)
		}
	}
}

func TaskRspamdCheck() {

	result := db.MailListForRspamd(1)
	rspamdEnable, _ := config.GetBool("rspamd.enable", false)
	if rspamdEnable {
		for _, val := range result {
			_, err, score := libs.RspamdCheck(val.Content)
			// fmt.Println("RspamdCheck:", val.Id, err)
			if err == nil {
				db.MailSetIsCheckById(val.Id, 1)
				log.Infof("mail[%d] check is pass! score:%f", val.Id, score)
			} else {
				db.MailSetIsCheckById(val.Id, 1)
				db.MailSetJunkById(val.Id, 1)
				log.Errorf("mail[%d] check is spam! score:%f", val.Id, score)
			}
		}
	}
}

func Init() {
	c := cron.New()

	c.AddFunc("*/5 * * * * * ", func() {
		// fmt.Println(fmt.Sprintf("TaskQueueeSendMail! time:%d", time.Now().Unix()))
		TaskQueueeSendMail()
	})

	c.AddFunc("*/10 * * * * * ", func() {
		TaskRspamdCheck()
	})

	c.Start()
}
