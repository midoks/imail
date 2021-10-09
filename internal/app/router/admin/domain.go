package admin

import (
	// "fmt"
	// "errors"

	"github.com/midoks/imail/internal/app/context"
	"github.com/midoks/imail/internal/app/form"
	"github.com/midoks/imail/internal/conf"
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

	d, _ := db.DomainList(1, 10)

	c.Data["Total"] = db.DomainCount()
	c.Data["Domain"] = d
	c.Success(DOMAIN)
}

func NewDomain(c *context.Context) {
	c.Data["Title"] = c.Tr("admin.domain")
	c.Data["PageIsAdmin"] = true
	c.Data["PageIsAdminDomain"] = true

	c.Success(DOMAIN_NEW)
}

func NewDomainPost(c *context.Context, f form.AdminCreateDomain) {
	c.Data["Title"] = c.Tr("admin.domain")
	c.Data["PageIsAdmin"] = true
	c.Data["PageIsAdminDomain"] = true
	count := db.DomainCount()

	limit := 9
	if int(count) >= limit {
		c.FormErr("Domain")
		c.RenderWithErr(c.Tr("form.domain_add_limit_exceeded", limit), DOMAIN_NEW, &f)
		return
	}

	if c.HasError() {
		c.Success(DOMAIN_NEW)
		return
	}

	d := &db.Domain{
		Domain: f.Domain,
	}

	err := db.DomainCreate(d)
	if err != nil {
		c.FormErr("Domain")
		c.RenderWithErr(c.Tr("admin.domain.add_fail", f.Domain), DOMAIN_NEW, &f)
		return
	}

	c.Flash.Success(c.Tr("admin.domain.add_success", f.Domain))
	c.Redirect(conf.Web.Subpath + "/admin/domain")
}

func DeleteDomain(c *context.Context) {
	id := c.ParamsInt64(":id")
	err := db.DomainDeleteById(id)
	if err != nil {
		c.Flash.Success(c.Tr("admin.domain.deletion_fail"))
	} else {
		c.Flash.Success(c.Tr("admin.domain.deletion_success"))
	}
	c.Redirect(conf.Web.Subpath + "/admin/domain")
}
