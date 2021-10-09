package db

import (
	"time"
)

type Domain struct {
	Id          int64     `gorm:"primaryKey"`
	Domain      string    `gorm:"unique;comment:顶级域名"`
	Mx          bool      `gorm:"comment:MX记录"`
	A           bool      `gorm:"comment:A记录"`
	Spf         bool      `gorm:"comment:Spf记录"`
	Dkim        bool      `gorm:"comment:Dkim记录"`
	Dmarc       bool      `gorm:"comment:DMARC记录"`
	IsDefault   bool      `gorm:"comment:是否默认"`
	Created     time.Time `gorm:"autoCreateTime;comment:创建时间"`
	CreatedUnix int64     `gorm:"autoCreateTime;comment:创建时间"`
	Updated     time.Time `gorm:"autoCreateTime;comment:更新时间"`
	UpdatedUnix int64     `gorm:"autoCreateTime;comment:更新时间"`
}

func DomainTableName() string {
	return "im_domain"
}

func (*Domain) TableName() string {
	return DomainTableName()
}

func DomainCreate(d *Domain) (err error) {
	data := db.First(d, "domain = ?", d.Domain)
	if data.Error != nil {
		result := db.Create(d)
		return result.Error
	}
	return data.Error
}

func DomainCount() int64 {
	var count int64
	db.Model(&Domain{}).Count(&count)
	return count
}

func DomainList(page, pageSize int) ([]*Domain, error) {
	domain := make([]*Domain, 0, pageSize)
	dbm := db.Limit(pageSize).Offset((page - 1) * pageSize).Order("id desc")
	err := dbm.Find(&domain)
	return domain, err.Error
}
