package db

import (
	// "fmt"
	// "github.com/midoks/imail/internal/libs"
	// "strings"
	"errors"

	"time"
)

type UserLoginVerify struct {
	Name       string `gorm:"size:50;comment:用户"`
	Rand       string `gorm:"size:50;comment:随机码"`
	Token      string `gorm:"size:50;comment:编码"`
	Expire     int    `gorm:"comment:状态"`
	UpdateTime int64  `gorm:"autoCreateTime;comment:更新时间"`
	CreateTime int64  `gorm:"autoCreateTime;comment:创建时间"`
}

func (UserLoginVerify) TableName() string {
	return "im_users_login_verify"
}

func UserLoginVerifyGet(name string) (UserLoginVerify, error) {
	var ulv UserLoginVerify
	db.Where("name = ?", name).First(&ulv)
	if ulv.Name == "" {
		return ulv, errors.New("record not found")
	}
	return ulv, nil
}

func UserLoginVerifyAdd(name string, rand string, token string) (int64, error) {

	_, err := UserLoginVerifyGet(name)
	if err != nil {
		u := UserLoginVerify{
			Name:  name,
			Rand:  rand,
			Token: token,
		}

		u.UpdateTime = time.Now().Unix()
		u.CreateTime = time.Now().Unix()
		result := db.Create(&u)
		return result.RowsAffected, result.Error
	} else {
		db.Model(&UserLoginVerify{}).Where("name = ?", name).Update("rand", rand).Update("token", token)
	}
	return 1, nil
}
