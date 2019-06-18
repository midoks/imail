package models

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	"time"
)

type Class struct {
	Id         int64
	Name       string `orm:"size(50);comment(分类名)"`
	Type       string `orm:"size(50);comment(类型)"`
	Userid     int64  `orm:"comment(用户ID)"`
	UpdateTime int64
	CreateTime int64
}

func ClassTableName() string {
	return "im_class"
}

func (u *Class) TableName() string {
	return ClassTableName()
}

func (u *Class) Update(fields ...string) error {
	u.UpdateTime = time.Now().Unix()
	if _, err := orm.NewOrm().Update(u, fields...); err != nil {
		return err
	}
	return nil
}

func ClassGetByUid(uid int64) ([]orm.Params, error) {
	var maps []orm.Params

	o := orm.NewOrm()
	sql := fmt.Sprintf("SELECT name,tag FROM `%s` WHERE (`type`=0) or (uid=?)", ClassTableName())
	num, err := o.Raw(sql, uid).Values(&maps)
	if err == nil && num > 0 {

		return maps, nil
	}
	return maps, err
}
