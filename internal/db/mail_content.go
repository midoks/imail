package db

import (
	// "fmt"
	"strings"

	"github.com/midoks/imail/internal/conf"
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
	mode := conf.Web.MailSaveMode
	if strings.EqualFold(mode, "hard_disk") {
		return MailContentReadHardDisk(mid)
	} else {
		return MailContentReadDb(mid)
	}
}

func MailContentReadDb(mid int64) (string, error) {
	var r MailContent
	err := db.Model(&MailContent{}).Where("mid = ?", mid).First(&r).Error
	if err != nil {
		return "", err
	}
	return r.Content, nil
}

func MailContentReadHardDisk(mid int64) (string, error) {
	var r MailContent
	err := db.Model(&MailContent{}).Where("mid = ?", mid).First(&r).Error
	if err != nil {
		return "", err
	}
	return r.Content, nil
}

func MailContentWrite(uid int64, mid int64, content string) error {
	mode := conf.Web.MailSaveMode
	if strings.EqualFold(mode, "hard_disk") {
		return MailContentWriteHardDisk(uid, mid, content)
	} else {
		return MailContentWriteDb(mid, content)
	}
}

func MailContentWriteDb(mid int64, content string) error {
	user := MailContent{Mid: mid, Content: content}
	result := db.Create(&user)
	return result.Error
}

func MailContentWriteHardDisk(uid int64, mid int64, content string) error {
	user := MailContent{Mid: mid, Content: content}
	result := db.Create(&user)
	return result.Error
}
