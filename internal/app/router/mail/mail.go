package mail

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/midoks/imail/internal/app/context"
	"github.com/midoks/imail/internal/app/form"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/tools"
	tmail "github.com/midoks/imail/internal/tools/mail"
	"github.com/midoks/imail/internal/tools/paginater"
)

const (
	MAIL        = "mail/list"
	MAIL_NEW    = "mail/new"
	MAIL_CONENT = "mail/content"
)

type MailSearchOptions struct {
	page     int
	PageSize int
	OrderBy  string
	TplName  string
	Type     int
	Bid      int64
}

func RenderMailSearch(c *context.Context, opts *MailSearchOptions) {
	page := c.QueryInt("page")
	if page <= 1 {
		page = 1
	}

	var (
		mail  []*db.Mail
		count int64
		err   error
	)

	keyword := c.Query("q")
	opt := &db.MailSearchOptions{
		Keyword:  keyword,
		OrderBy:  opts.OrderBy,
		Page:     page,
		PageSize: opts.PageSize,
		Type:     opts.Type,
		Uid:      c.User.Id,
	}

	if len(keyword) == 0 {
		mail, err = db.MailList(page, opts.PageSize, opt)
		count = db.MailCountWithOpts(opt)
	} else {
		mail, count, err = db.MailSearchByName(opt)
		if err != nil {
			c.Error(err, "search user by name")
			return
		}
	}
	c.Data["Keyword"] = keyword
	c.Data["Total"] = count
	c.Data["Bid"] = opts.Bid
	c.Data["Page"] = paginater.New(int(count), opts.PageSize, page, 5)
	c.Data["Mail"] = mail

	c.Success(opts.TplName)
}

func Flags(c *context.Context) {
	c.Data["Title"] = c.Tr("mail.flags")
	c.Data["PageIsMail"] = true

	bid := c.ParamsInt64(":bid")
	RenderMailSearch(c, &MailSearchOptions{
		PageSize: 10,
		OrderBy:  "id ASC",
		TplName:  MAIL,
		Type:     db.MailSearchOptionsTypeFlags,

		Bid: bid,
	})
}

func Sent(c *context.Context) {
	c.Data["Title"] = c.Tr("mail.sent")
	c.Data["PageIsMail"] = true

	bid := c.ParamsInt64(":bid")

	RenderMailSearch(c, &MailSearchOptions{
		PageSize: 10,
		OrderBy:  "id ASC",
		TplName:  MAIL,
		Type:     db.MailSearchOptionsTypeSend,
		Bid:      bid,
	})
}

func Deleted(c *context.Context) {
	c.Data["Title"] = c.Tr("mail.deleted")
	c.Data["PageIsMail"] = true

	bid := c.ParamsInt64(":bid")

	RenderMailSearch(c, &MailSearchOptions{
		PageSize: 10,
		OrderBy:  "id ASC",
		TplName:  MAIL,
		Type:     db.MailSearchOptionsTypeDeleted,
		Bid:      bid,
	})
}

func Junk(c *context.Context) {
	c.Data["Title"] = c.Tr("mail.junk")
	c.Data["PageIsMail"] = true

	bid := c.ParamsInt64(":bid")

	RenderMailSearch(c, &MailSearchOptions{
		PageSize: 10,
		OrderBy:  "id ASC",
		TplName:  MAIL,
		Type:     db.MailSearchOptionsTypeJunk,
		Bid:      bid,
	})
}

func Mail(c *context.Context) {

	c.Data["Title"] = c.Tr("mail.write_letter")
	c.Data["PageIsMail"] = true

	// c.Success(MAIL)
	bid := c.ParamsInt64(":bid")
	RenderMailSearch(c, &MailSearchOptions{
		PageSize: 10,
		OrderBy:  "id ASC",
		TplName:  MAIL,
		Type:     db.MailSearchOptionsTypeInbox,
		Bid:      bid,
	})
}

func New(c *context.Context) {
	c.Data["Title"] = c.Tr("mail.write_letter")
	c.Data["PageIsWriteMail"] = true
	bid := c.ParamsInt64(":bid")
	c.Data["Bid"] = bid

	c.Success(MAIL_NEW)
}

func NewPost(c *context.Context, f form.SendMail) {
	c.Data["Title"] = c.Tr("mail.write_letter")
	c.Data["PageIsWriteMail"] = true

	bid := c.ParamsInt64(":bid")
	c.Data["Bid"] = bid

	from, err := db.DomainGetMainForDomain()
	if err != nil {
		c.RenderWithErr(c.Tr("mail.new.default_not_set"), MAIL_NEW, &f)
		return
	}

	mail_from := fmt.Sprintf("%s@%s", c.User.Name, from)
	tc, err := tmail.GetMailReturnToSender(mail_from, f.ToMail, f.Subject, f.Content)

	_, err = db.MailPushSend(c.User.Id, mail_from, f.ToMail, tc)
	if err != nil {
		c.RenderWithErr(err.Error(), MAIL_NEW, &f)
		return
	}
	c.Flash.Success("OK")
	c.Success(MAIL_NEW)
}

func Content(c *context.Context) {
	c.Data["Title"] = c.Tr("mail.write_letter")
	c.Data["PageIsMailContent"] = true

	id := c.ParamsInt64(":id")
	c.Data["id"] = id

	bid := c.ParamsInt64(":bid")
	c.Data["Bid"] = bid

	r, err := db.MailById(id)
	if err == nil {
		c.Data["Mail"] = r
	}

	c.Success(MAIL_CONENT)
}

/**
 * API
 **/
func ApiRead(c *context.Context, f form.MailIDs) {
	ids := f.Ids
	idsSlice, err := tools.ToSlice(ids)
	if err != nil {
		c.Fail(-1, c.Tr("common.fail"))
		return
	}

	if db.MailSeenByIds(idsSlice) {
		c.OK(c.Tr("common.success"))
	} else {
		c.Fail(-1, c.Tr("common.fail"))
	}
}

func ApiUnread(c *context.Context, f form.MailIDs) {
	ids := f.Ids
	idsSlice, err := tools.ToSlice(ids)
	if err != nil {
		c.Fail(-1, c.Tr("common.fail"))
		return
	}

	if db.MailUnSeenByIds(idsSlice) {
		c.OK(c.Tr("common.success"))
	} else {
		c.Fail(-1, c.Tr("common.fail"))
	}
}

func ApiStar(c *context.Context, f form.MailIDs) {
	int64, _ := strconv.ParseInt(f.Ids, 10, 64)
	if db.MailSetFlagsById(int64, 1) {
		c.OK(c.Tr("common.success"))
	} else {
		c.Fail(-1, c.Tr("common.fail"))
	}
}

func ApiUnStar(c *context.Context, f form.MailIDs) {
	int64, _ := strconv.ParseInt(f.Ids, 10, 64)
	if db.MailSetFlagsById(int64, 0) {
		c.OK(c.Tr("common.success"))
	} else {
		c.Fail(-1, c.Tr("common.fail"))
	}
}

func ApiMove(c *context.Context, f form.MailIDs) {
	ids := f.Ids
	dir := f.Dir

	idsSlice, err := tools.ToSlice(ids)
	if err != nil {
		c.Fail(-1, c.Tr("common.fail"))
		return
	}

	if strings.EqualFold(dir, "deleted") {
		if db.MailSoftDeleteByIds(idsSlice) {
			c.OK(c.Tr("common.success"))
			return
		}
	}

	if strings.EqualFold(dir, "junk") {
		if db.MailSetJunkByIds(idsSlice, 1) {
			c.OK(c.Tr("common.success"))
			return
		}
	}

	c.Fail(-1, c.Tr("common.fail"))
	return
}
