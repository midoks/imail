package task

import (
	"fmt"
	"github.com/robfig/cron"
	"time"
)

func Init() {
	c := cron.New()

	c.AddFunc("*/5 * * * * * ", func() {
		fmt.Println(fmt.Sprintf("task test! time:%d", time.Now().Unix()))
	})

	c.Start()
}
