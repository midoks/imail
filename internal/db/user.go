package db

import (
	"fmt"
	_ "github.com/midoks/imail/internal/libs"
	// "gorm.io/gorm"
	"strings"
	_ "time"
)

type User struct {
	Id         int64  `gorm:"primaryKey"`
	Name       string `gorm:"unique;size:50;comment:用户名"`
	Password   string `gorm:"size:32;comment:用户密码"`
	Code       string `gorm:"size:50;comment:编码"`
	Role       int    `gorm:"comment:角色"`
	Status     int    `gorm:"comment:状态"`
	UpdateTime int64  `gorm:"autoCreateTime;comment:更新时间"`
	CreateTime int64  `gorm:"autoCreateTime;comment:创建时间"`
}

func (User) TableName() string {
	return "im_users"
}

func LoginWithCode(name string, code string) (bool, int64) {
	list := strings.SplitN(name, "@", 2)
	var user User
	err := db.First(&user, "name = ?", list[0]).Error
	if err != nil {
		return false, 0
	}
	fmt.Println(user)
	return true, user.Id
}

func UserGetByName(name string) (User, error) {
	list := strings.SplitN(name, "@", 2)
	var user User
	err := db.First(&user, "name = ?", list[0]).Error
	if err != nil {
		return user, err
	}
	return user, nil
}
