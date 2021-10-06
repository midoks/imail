package mail

import (
	"github.com/midoks/imail/internal/app/context"
)

const (
	MAIL     = "mail/list"
	USER_NEW = "mail/new"
)

func Mail(c *context.Context) {
	c.Data["Title"] = c.Tr("mail.write_letter")
	c.Data["PageIsWriteMail"] = true

	c.Success(MAIL)
}

func New(c *context.Context) {
	c.Data["Title"] = c.Tr("mail.write_letter")
	c.Data["PageIsWriteMail"] = true

	c.Success(USER_NEW)
}
