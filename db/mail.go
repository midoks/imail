package db

import (
	"fmt"
	_ "github.com/midoks/imail/libs"
	// "gorm.io/gorm"
	// "strings"
	"errors"
	"time"
)

type Mail struct {
	Id         int64  `gorm:"primaryKey"`
	MailFrom   string `gorm:"size:50;comment:邮件来源"`
	MailTo     string `gorm:"size:50;comment:接收邮件"`
	Content    string `gorm:"comment:邮件内容"`
	Size       int    `gorm:"size:50;comment:邮件内容大小"`
	Status     int
	UpdateTime int64 `gorm:"autoCreateTime;comment:更新时间"`
	CreateTime int64 `gorm:"autoCreateTime;comment:创建时间"`
}

func (Mail) TableName() string {
	return "im_mail"
}

func MailPush(mail_from string, mail_to string, content string) (int64, error) {

	user := Mail{
		MailFrom: mail_from,
		MailTo:   mail_to,
		Content:  content,
		Size:     len(content),
	}

	user.UpdateTime = time.Now().Unix()
	user.CreateTime = time.Now().Unix()
	result := db.Create(&user)

	fmt.Println(result)

	return 0, errors.New("error")
}
