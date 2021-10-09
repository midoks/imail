package admin

import (
	"fmt"
	"net"
	"strings"
	// "errors"

	"github.com/midoks/imail/internal/app/context"
	"github.com/midoks/imail/internal/app/form"
	"github.com/midoks/imail/internal/conf"
	"github.com/midoks/imail/internal/db"
	// "github.com/midoks/imail/internal/tools"
	"github.com/midoks/imail/internal/tools/dkim"
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

func CheckDomain(c *context.Context) {
	id := c.ParamsInt64(":id")
	d, _ := db.DomainGetById(id)
	domain := d.Domain

	//MX
	mx, _ := net.LookupMX(domain)
	lenMx := len(mx)
	if 0 == lenMx {
		d.Mx = false
	} else {
		if strings.Contains(mx[0].Host, ".") {
			d.Mx = true
		}
	}

	//A
	d.A = false
	err := dkim.CheckDomainA(domain)
	if err == nil {
		d.A = true
	}

	//DMARC
	dmarcRecord, _ := net.LookupTXT(fmt.Sprintf("_dmarc.%s", domain))
	if 0 == len(dmarcRecord) {
		d.Dmarc = false
	} else {
		for _, dmarcDomainRecord := range dmarcRecord {
			if strings.Contains(strings.ToLower(dmarcDomainRecord), "v=dmarc1") {
				d.Dmarc = true
			} else {
				d.Dmarc = false
			}
		}
	}

	//spf
	spfRecord, _ := net.LookupTXT(domain)
	if 0 == len(spfRecord) {
		d.Spf = false
	} else {
		for _, spfRecordContent := range spfRecord {
			if strings.Contains(strings.ToLower(spfRecordContent), "v=spf1") {
				d.Spf = true
			}
		}
	}

	//dkim check
	dataDir := conf.Web.Subpath + conf.Web.AppDataPath
	d.Dkim = false
	dkimRecord, _ := net.LookupTXT(fmt.Sprintf("default._domainkey.%s", domain))
	fmt.Println("dkimRecord:", dkimRecord, domain)
	if 0 != len(dkimRecord) {
		dkimContent, _ := dkim.GetDomainDkimVal(dataDir, domain)
		for _, dkimDomainContent := range dkimRecord {
			if strings.EqualFold(dkimContent, dkimDomainContent) {
				d.Dkim = true
			}
		}
	}

	_ = db.DomainUpdateById(id, d)

	c.Flash.Success(c.Tr("admin.domain.check_success", d.Domain))
	c.Redirect(conf.Web.Subpath + "/admin/domain")
}

func SetDefaultDomain(c *context.Context) {
	id := c.ParamsInt64(":id")
	d, _ := db.DomainGetById(id)
	err := db.DomainSetDefaultOnlyOne(id)
	fmt.Println("SetDefaultDomain:", err)
	if err != nil {
		c.Flash.Error(c.Tr("admin.domain.set_default_fail", d.Domain))
	} else {
		c.Flash.Success(c.Tr("admin.domain.set_default_success", d.Domain))
	}
	c.Redirect(conf.Web.Subpath + "/admin/domain")
}
