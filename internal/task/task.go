package task

import (
	"fmt"
	"github.com/midoks/imail/internal/config"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/smtpd"
	"github.com/robfig/cron"
	"sync"
	"time"
)

func TaskQueueeSendMail() {
	var wg sync.WaitGroup
	result := db.MailSendListForStatus(1)

	for i, val := range result {

		wg.Add(i)
		err := smtpd.Delivery("", val.MailFrom, val.MailTo, []byte(val.Content))
		if err == nil {
			db.MailSetStatusById(val.Id, 1)
		} else {
			domain := config.GetString("mail.domain", "xxx.com")
			postmaster := fmt.Sprintf("postmaster@%s", domain)
			err = smtpd.Delivery("", postmaster, val.MailTo, []byte("系统退信息"))
			fmt.Println("ddd:", err)
		}
		fmt.Println("send mail:", err)
		wg.Wait()
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
