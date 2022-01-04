package db

import (
	_ "fmt"
)

type MailContent struct {
	Id      int64  `gorm:"primaryKey"`
	Mid     int64  `gorm:"index:uniqueIndex;comment:MID"`
	Content string `gorm:"comment:Content"`
}

func MailContentTableName() string {
	return "im_mail_content"
}

func (*MailContent) TableName() string {
	return MailContentTableName()
}

func MailContentRead(mid int64) (string, error) {
	var r MailContent
	err := db.Model(&MailContent{}).Where("mid = ?", mid).First(&r).Error
	if err != nil {
		return "", err
	}
	return r.Content, nil
}

func MailContentWrite(mid int64, content string) error {
	user := MailContent{Mid: mid, Content: content}
	result := db.Create(&user)
	return result.Error
}
