package models

import (
	_ "fmt"
	"time"

	"github.com/astaxie/beego/orm"
)

type UserMailBox struct {
	Id         int64
	Uid        int64 `comment(用户ID)"`
	Mid        int64 `comment(邮件ID)"`
	UpdateTime int64
	CreateTime int64
}

func (u *UserMailBox) TableName() string {
	return "im_user_box"
}

func (u *UserMailBox) Update(fields ...string) error {
	u.UpdateTime = time.Now().Unix()
	if _, err := orm.NewOrm().Update(u, fields...); err != nil {
		return err
	}
	return nil
}

func BoxAdd(uid int64, mid int64) (int64, error) {
	data := new(UserMailBox)
	data.Uid = uid
	data.Mid = mid

	data.UpdateTime = time.Now().Unix()
	data.CreateTime = time.Now().Unix()
	i, err := orm.NewOrm().Insert(data)
	if err != nil {
		return 0, err
	}
	return i, err
}
