package models

import (
	_ "fmt"
	"github.com/astaxie/beego/orm"
	"time"
)

type User struct {
	Id         int
	Name       string `json:"name";orm:"unique;size(50);comment(用户名)"`
	Password   string `json:"password";orm:"unique;size(50);comment(用户密码)"`
	UpdateTime int64
	CreateTime int64
}

func getTnByUser() string {
	return "im_users"
}

func (u *User) TableName() string {
	return getTnByUser()
}

func (u *User) Update(fields ...string) error {
	u.UpdateTime = time.Now().Unix()
	if _, err := orm.NewOrm().Update(u, fields...); err != nil {
		return err
	}
	return nil
}

func UserGetById(id int) (*User, error) {
	u := new(User)
	err := orm.NewOrm().QueryTable(getTnByUser()).Filter("id", id).One(u)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// func UserGetByName(username string) (*ImUser, error) {

// 	u := new(ImUser)
// 	err := orm.NewOrm().QueryTable(getTnByUser()).Filter("username", username).One(u)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return u, nil
// }

// func UserDelById(id int) (int64, error) {
// 	return orm.NewOrm().Delete(&ImUser{Id: id})
// }
