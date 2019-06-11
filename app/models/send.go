package models

import (
	_ "fmt"
	"time"

	"github.com/astaxie/beego/orm"
)

type UserMailSend struct {
	Id         int64  `orm:"comment(ID)"`
	Uid        int64  `orm:"comment(用户ID)"`
	MailFrom   string `orm:"comment(ID)"`
	MailTo     string `orm:"comment(邮件ID)"`
	Content    string `orm:"comment(邮件内容)"`
	Status     int    `orm:"comment(邮件ID)"`
	UpdateTime int64  `orm:"comment(更新时间)"`
	CreateTime int64  `orm:"comment(创建时间)"`
}

func (u *UserMailSend) TableName() string {
	return "im_user_send"
}

func (u *UserMailSend) Update(fields ...string) error {
	u.UpdateTime = time.Now().Unix()
	if _, err := orm.NewOrm().Update(u, fields...); err != nil {
		return err
	}
	return nil
}

func SendAdd(uid int64, mail_from string, mail_to string, content string) (int64, error) {
	data := new(UserMailSend)
	data.Uid = uid
	data.MailFrom = mail_from
	data.MailTo = mail_to
	data.Content = content

	data.UpdateTime = time.Now().Unix()
	data.CreateTime = time.Now().Unix()
	i, err := orm.NewOrm().Insert(data)
	if err != nil {
		return 0, err
	}
	return i, err
}
