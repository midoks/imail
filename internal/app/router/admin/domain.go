package admin

import (
	"github.com/midoks/imail/internal/app/context"
	"github.com/midoks/imail/internal/db"
)

const (
	DOMAIN     = "admin/domain/list"
	DOMAIN_NEW = "admin/domain/new"
)

func Domain(c *context.Context) {
	c.Data["Title"] = c.Tr("admin.domain")
	c.Data["PageIsAdmin"] = true
	c.Data["PageIsAdminDomain"] = true

	c.Data["Total"] = db.DomainCount()
	c.Success(DOMAIN)
}

func NewDomain(c *context.Context) {
	c.Data["Title"] = c.Tr("admin.domain")
	c.Data["PageIsAdmin"] = true
	c.Data["PageIsAdminDomain"] = true

	c.Data["Total"] = db.DomainCount()
	c.Success(DOMAIN_NEW)
}
