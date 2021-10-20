package task

import (
	"fmt"
	"github.com/midoks/imail/internal/conf"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/log"
	"github.com/midoks/imail/internal/smtpd"
	"github.com/midoks/imail/internal/tools/cron"
	"github.com/midoks/imail/internal/tools/mail"
)

var c = cron.New()

func TaskQueueeSendMail() {
	postmaster := fmt.Sprintf("postmaster@%s", conf.Web.Domain)

	result := db.MailSendListForStatus(2, 1)
	if len(result) == 0 {

		result = db.MailSendListForStatus(0, 1)
		fmt.Println(result)
		for _, val := range result {
			db.MailSetStatusById(val.Id, 2)
			err := smtpd.Delivery("", val.MailFrom, val.MailTo, []byte(val.Content))
			if err != nil {

				content, _ := mail.GetMailReturnToSender(val.MailFrom, val.MailTo, val.Content, err.Error())
				db.MailPush(val.Uid, 1, postmaster, val.MailFrom, content, 1)
			}
			db.MailSetStatusById(val.Id, 1)
			fmt.Println("send mail:", err)
		}
	}
}

func TaskRspamdCheck() {

	result := db.MailListForRspamd(1)
	if conf.Rspamd.Enable {
		for _, val := range result {
			_, err, score := mail.RspamdCheck(val.Content)
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

	c.AddFunc("mail send task", "@every 5s", func() { TaskQueueeSendMail() })
	c.AddFunc("mail rspamd check", "@every 10m", func() { TaskRspamdCheck() })

	c.Start()
}

// ListTasks returns all running cron tasks.
func ListTasks() []*cron.Entry {
	return c.Entries()
}
