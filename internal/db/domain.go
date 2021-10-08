package db

import (
	"time"
)

type Domain struct {
	Id     int64  `gorm:"primaryKey"`
	Doamin string `gorm:"comment:顶级域名"`
	Mx     bool   `gorm:"comment:MX记录"`
	A      bool   `gorm:"comment:A记录"`
	Spf    bool   `gorm:"comment:Spf记录"`
	Dkim   bool   `gorm:"comment:Dkim记录"`
	Dmarc  bool   `gorm:"comment:DMARC记录"`

	IsDefalut bool `gorm:"comment:是否默认"`

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
