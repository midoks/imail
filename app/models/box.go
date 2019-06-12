package models

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	"time"
)

type UserMailBox struct {
	Id         int64
	Uid        int64 `comment(用户ID)"`
	Mid        int64 `comment(邮件ID)"`
	Type       int   `comment(类型|0:接收邮件;1:发送邮件)`
	Size       int   `comment(邮件内容大小[byte])`
	UpdateTime int64
	CreateTime int64
}

func BoxTableName() string {
	return "im_user_box"
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

func BoxAdd(uid int64, mid int64, method int, size int) (int64, error) {
	data := new(UserMailBox)
	data.Uid = uid
	data.Mid = mid
	data.Size = size
	data.Type = method

	data.UpdateTime = time.Now().Unix()
	data.CreateTime = time.Now().Unix()
	i, err := orm.NewOrm().Insert(data)
	if err != nil {
		return 0, err
	}
	return i, err
}

func BoxUserTotal(uid int64) (int64, int64) {
	fmt.Println("uid", uid)

	type User struct {
		Count int64
		Size  int64
	}
	var user User

	o := orm.NewOrm()
	err := o.Raw("SELECT count(uid) as count, sum(size) as size FROM `im_user_box` WHERE uid=?", uid).QueryRow(&user)
	if err != nil {

	}
	fmt.Println(user, err)

	return 0, 0
}

func BoxList(uid int64, page int, pageSize int) ([]*UserMailBox, int64) {

	offset := (page - 1) * pageSize
	list := make([]*UserMailBox, 0)

	query := orm.NewOrm().QueryTable(BoxTableName())
	total, _ := query.Count()
	query.Limit(pageSize, offset).All(&list)
	return list, total
}
