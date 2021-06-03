package db

import (
	"fmt"
	_ "github.com/midoks/imail/libs"
	// "gorm.io/gorm"
	// "errors"
	// "strings"
	// "time"
)

type Class struct {
	Id         int64  `gorm:"primaryKey"`
	Name       string `gorm:"size:50;comment:分类名"`
	Type       string `gorm:"size:50;comment:类型"`
	Userid     int64  `gorm:"comment:用户ID"`
	UpdateTime int64  `gorm:"autoCreateTime;comment:更新时间"`
	CreateTime int64  `gorm:"autoCreateTime;comment:创建时间"`
}

func ClassTableName() string {
	return "im_class"
}

func (Class) TableName() string {
	return ClassTableName()
}

// func (u *Class) Update(fields ...string) error {
// 	u.UpdateTime = time.Now().Unix()
// 	if _, err := orm.NewOrm().Update(u, fields...); err != nil {
// 		return err
// 	}
// 	return nil
// }

// func ClassGetByUid(uid int64) ([]orm.Params, error) {
// 	var maps []orm.Params

// 	o := orm.NewOrm()
// 	sql := fmt.Sprintf("SELECT name,flags FROM `%s` WHERE (`type`=0) or (uid=?)", ClassTableName())
// 	num, err := o.Raw(sql, uid).Values(&maps)
// 	if err == nil && num > 0 {

// 		return maps, nil
// 	}
// 	return maps, err
// }

func ClassGetIdByName(uid int64, name string) (int64, error) {
	var class Class
	sql := fmt.Sprintf("SELECT id,name,flags FROM `%s` WHERE (`type`=0 or uid=?) and `name`=?", ClassTableName())
	result := db.Raw(sql, uid, name).Find(&class)

	fmt.Println(result)

	// if err == nil && num > 0 {
	// 	id, err := strconv.ParseInt(class.Id, 10, 64)
	// 	if err == nil {
	// 		return id, nil
	// 	}
	// 	return id, err
	// }
	return 0, nil
}
