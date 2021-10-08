package admin

import (
	"github.com/midoks/imail/internal/app/context"
	"github.com/midoks/imail/internal/db"
)

const (
	DOMAIN = "admin/domain/list"
)

func Domain(c *context.Context) {
	c.Data["Title"] = c.Tr("admin.domain")
	c.Data["PageIsAdmin"] = true
	c.Data["PageIsAdminDomain"] = true

	c.Data["Total"] = db.DomainCount()
	c.Success(DOMAIN)
}
