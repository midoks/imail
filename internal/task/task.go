package task

import (
	"fmt"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/smtpd"
	"github.com/robfig/cron"
	"time"
)

func TaskQueueeSendMail() {
	result := db.MailSendListForStatus(1)
	// fmt.Println(result)
	for _, val := range result {
		err := smtpd.Delivery("", val.MailFrom, val.MailTo, []byte(val.Content))
		if err == nil {
			db.MailSetStatusById(val.Id, 1)
		}
		fmt.Println("send mail:", err)
	}

}

func Init() {
	c := cron.New()

	c.AddFunc("*/5 * * * * * ", func() {
		TaskQueueeSendMail()
		fmt.Println(fmt.Sprintf("TaskQueueeSendMail! time:%d", time.Now().Unix()))
	})

	c.Start()
}
