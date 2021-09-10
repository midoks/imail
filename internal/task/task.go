package task

import (
	"fmt"
	"github.com/Shopify/go-rspamd"
	"github.com/midoks/imail/internal/config"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/libs"
	"github.com/midoks/imail/internal/smtpd"
	"github.com/robfig/cron"
	// "sync"
	"bytes"

	"context"
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

	for _, val := range result {

		client := rspamd.New("http://127.0.0.1:11334")
		// client := rspamd.New("http://rspamd.cachecha.com")
		// client := rspamd.New("http://rspamd.cachecha.com", rspamd.Credentials("", "admin"))
		pong, err := client.Ping(context.Background())
		// fmt.Println("ddd:", pong, err, val)
		if err == nil {

			f := bytes.NewBuffer([]byte(val.Content))
			email := rspamd.NewEmailFromReader(f)
			checkRes, _ := client.Check(context.Background(), email)
			fmt.Println(checkRes)
			// fmt.Println(checkRes.MessageID)
			for _, symVal := range checkRes.Symbols {

				if symVal.Score > 0 {
					fmt.Println(symVal.Name, symVal.Score, symVal.Description)
				}
			}
			fmt.Println("mail[", val.Id, "] Score:", checkRes.Score)
		} else {
			fmt.Println(pong, err)
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
