package db

import (
// "errors"
// "fmt"
// "time"
)

type Queue struct {
	Id         int64  `gorm:"primaryKey"`
	Uid        int64  `gorm:"comment:用户ID"`
	Type       int    `gorm:"comment:0:发送;1:收到"`
	MailFrom   string `gorm:"size:50;comment:邮件来源"`
	MailTo     string `gorm:"size:50;comment:接收邮件"`
	Content    string `gorm:"comment:邮件内容"`
	Size       int    `gorm:"size:50;comment:邮件内容大小"`
	Status     int    `gorm:"comment:0:准备发送;1:发送成功;2:发送失败;3:已接收"`
	UpdateTime int64  `gorm:"autoCreateTime;comment:更新时间"`
	CreateTime int64  `gorm:"autoCreateTime;comment:创建时间"`
}

func (Queue) TableName() string {
	return TablePrefix("queue")
}
