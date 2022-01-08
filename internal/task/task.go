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

	from, err := db.DomainGetMainForDomain()
	if err != nil {
		return
	}

	postmaster := fmt.Sprintf("postmaster@%s", from)
	result := db.MailSendListForStatus(2, 1)
	if len(result) == 0 {

		result = db.MailSendListForStatus(0, 1)
		for _, val := range result {
			db.MailSetStatusById(val.Id, 2)

			content, err := db.MailContentRead(result[0].Uid, result[0].Id)
			if err != nil {
				continue
			}
			err = smtpd.Delivery("", val.MailFrom, val.MailTo, []byte(content))

			if err != nil {
				content, _ := mail.GetMailReturnToSender(postmaster, val.MailFrom, val.MailTo, content, err.Error())
				db.MailPush(val.Uid, 1, postmaster, val.MailFrom, content, 1)
			}
			db.MailSetStatusById(val.Id, 1)
		}
	}
}

func TaskRspamdCheck() {

	result := db.MailListForRspamd(1)
	if conf.Rspamd.Enable {
		for _, val := range result {
			content, err := db.MailContentRead(val.Uid, val.Id)
			if err != nil {
				continue
			}
			_, err, score := mail.RspamdCheck(content)
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
