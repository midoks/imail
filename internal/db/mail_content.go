package db

type MailContent struct {
	Id      int64  `gorm:"primaryKey"`
	Mid     int64  `gorm:"comment:MID"`
	Content string `gorm:"comment:内容"`
}

func (*MailContent) TableName() string {
	return MailTableName()
}

func MailContentTableName() string {
	return "im_mail_content"
}
