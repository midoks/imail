package mail

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/midoks/imail/internal/app/context"
	"github.com/midoks/imail/internal/app/form"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/tools"
	tmail "github.com/midoks/imail/internal/tools/mail"
	"github.com/midoks/imail/internal/tools/paginater"
	"github.com/midoks/mcopa"
)

const (
	MAIL             = "mail/list"
	MAIL_NEW         = "mail/new"
	MAIL_CONENT      = "mail/content"
	MAIL_CONENT_HTML = "mail/content_html"
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

func Draft(c *context.Context) {
	c.Data["Title"] = c.Tr("mail.draft")
	c.Data["PageIsMail"] = true
	c.Data["PageIsMailDraft"] = true

	bid := c.ParamsInt64(":bid")

	RenderMailSearch(c, &MailSearchOptions{
		PageSize: 10,
		OrderBy:  "id ASC",
		TplName:  MAIL,
		Type:     db.MailSearchOptionsTypeDraft,
		Bid:      bid,
	})
}

func Deleted(c *context.Context) {
	c.Data["Title"] = c.Tr("mail.deleted")
	c.Data["PageIsMail"] = true
	c.Data["PageIsMailDeleted"] = true

	bid := c.ParamsInt64(":bid")

	RenderMailSearch(c, &MailSearchOptions{
		PageSize: 10,
		OrderBy:  "id ASC",
		TplName:  MAIL,
		Type:     db.MailSearchOptionsTypeDeleted,
		Bid:      bid,
	})
}

func HardDeleteDraftMail(c *context.Context) {
	id := c.ParamsInt64(":id")

	mail, _ := db.MailById(id)
	if !db.MailHardDeleteById(mail.Uid, mail.Id) {
		c.Flash.Success(c.Tr("mail.draft.deletion_fail"))
	} else {
		c.Flash.Success(c.Tr("mail.draft.deletion_success"))
	}
	c.Redirect("/mail/draft")
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
	id := c.ParamsInt64(":id")

	mail, _ := db.MailById(id)
	content, _ := db.MailContentRead(mail.Uid, mail.Id)
	email, _ := mcopa.Parse(bufio.NewReader(strings.NewReader(content)))

	if strings.EqualFold(email.TextBody, "") {
		content = email.HTMLBody
	} else {
		content = email.TextBody
	}

	c.Data["Bid"] = bid
	c.Data["id"] = id

	c.Data["Mail"] = mail
	c.Data["MailContent"] = content

	c.Data["EditorLang"] = tools.ToEditorLang(c.Data["NowLang"].(string))

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
	tc, err := tmail.GetMailSend(mail_from, f.ToMail, f.Subject, f.Content)

	if f.Id != 0 {
		_, err = db.MailUpdate(f.Id, c.User.Id, 0, mail_from, f.ToMail, tc, 0, false)
	} else {
		_, err = db.MailPushSend(c.User.Id, mail_from, f.ToMail, tc, false)
	}

	if err != nil {
		c.RenderWithErr(err.Error(), MAIL_NEW, &f)
		return
	}

	c.RedirectSubpath("/mail/sent")
}

func NewPostDraft(c *context.Context, f form.SendMail) {
	c.Data["Title"] = c.Tr("mail.write_letter")
	c.Data["PageIsWriteMail"] = true

	bid := c.ParamsInt64(":bid")
	c.Data["Bid"] = bid

	from, err := db.DomainGetMainForDomain()
	if err != nil {
		c.Fail(-1, c.Tr("common.fail"))
		return
	}

	mail_from := fmt.Sprintf("%s@%s", c.User.Name, from)
	tc, err := tmail.GetMailSend(mail_from, f.ToMail, f.Subject, f.Content)

	var mid int64
	if f.Id != 0 {
		mid, err = db.MailUpdate(f.Id, c.User.Id, 0, mail_from, f.ToMail, tc, 0, true)
	} else {
		mid, err = db.MailPushSend(c.User.Id, mail_from, f.ToMail, tc, true)
	}

	if err == nil {
		r := make(map[string]int64)
		r["id"] = mid

		c.OKDATA(c.Tr("common.success"), r)
		return
	}

	c.Fail(-1, c.Tr("common.fail"))
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

	contentData, err := db.MailContentRead(r.Uid, id)
	if err != nil {
		c.Fail(-1, err.Error())
		return
	}

	content := bufio.NewReader(strings.NewReader(contentData))
	email, err := mcopa.Parse(content)
	if err != nil {
		c.Fail(-1, err.Error())
		return
	}

	c.Data["ParseMail"] = email

	c.Success(MAIL_CONENT)
}

func ContentHtml(c *context.Context) {
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

	contentData, err := db.MailContentRead(r.Uid, id)
	if err != nil {
		c.Fail(-1, err.Error())
		return
	}

	content := bufio.NewReader(strings.NewReader(contentData))
	email, err := mcopa.Parse(content)
	if err != nil {
		c.Fail(-1, err.Error())
		return
	}
	c.Data["ParseMail"] = email

	c.Success(MAIL_CONENT_HTML)
}

func ContentDownload(c *context.Context) {
	c.Data["Title"] = c.Tr("mail.write_letter")
	c.Data["PageIsMailContent"] = true

	id := c.ParamsInt64(":id")
	c.Data["id"] = id

	r, err := db.MailById(id)

	if err != nil {
		return
	}
	emailFilePath := db.MailContentFilename(r.Uid, id)
	tmpEmailName := fmt.Sprintf("imail_%d.eml", id)
	c.ServeFile(emailFilePath, tmpEmailName)
}

func ContentAttach(c *context.Context) {
	c.Data["Title"] = c.Tr("mail.write_letter")
	c.Data["PageIsMailContent"] = true

	id := c.ParamsInt64(":id")
	c.Data["id"] = id

	aid := c.ParamsInt(":aid")

	bid := c.ParamsInt64(":bid")
	c.Data["Bid"] = bid

	r, err := db.MailById(id)
	if err == nil {
		c.Data["Mail"] = r
	}

	contentData, err := db.MailContentRead(r.Uid, id)
	if err != nil {
		c.Fail(-1, err.Error())
		return
	}

	content := bufio.NewReader(strings.NewReader(contentData))
	email, err := mcopa.Parse(content)
	if err != nil {
		c.Fail(-1, err.Error())
		return
	}
	c.Data["ParseMail"] = email

	attachFile, err := ioutil.ReadAll(email.Attachments[aid].Data)
	pathName := "/tmp/" + email.Attachments[aid].Filename
	tools.WriteFile(pathName, string(attachFile))

	c.ServeFile(pathName, email.Attachments[aid].Filename)
	os.RemoveAll(pathName)

	// return macaron.ReturnStruct{Code: http.StatusOK, Data: string(attachFile)}
}

func ContentDemo(c *context.Context) {

	id := c.ParamsInt64(":id")
	contentData, err := db.MailContentRead(1, id)
	if err != nil {
		c.Fail(-1, err.Error())
		return
	}

	bufferedBody := bufio.NewReader(strings.NewReader(contentData))
	email, err := mcopa.Parse(bufferedBody)
	if err != nil {
		c.Fail(-1, err.Error())
		return
	}

	c.OKDATA("ok", email)
}

/****************************************************
 * API for web frontend call
 ***************************************************/
func ApiDeleted(c *context.Context, f form.MailIDs) {
	ids := f.Ids
	idsSlice, err := tools.ToSlice(ids)
	if err != nil {
		c.Fail(-1, c.Tr("common.fail"))
		return
	}

	if db.MailSoftDeleteByIds(idsSlice) {
		c.OK(c.Tr("common.success"))
	} else {
		c.Fail(-1, c.Tr("common.fail"))
	}
}

//TODO:硬删除
func ApiHardDeleted(c *context.Context, f form.MailIDs) {
	ids := f.Ids
	idsSlice, err := tools.ToSlice(ids)
	if err != nil {
		c.Fail(-1, c.Tr("common.fail"))
		return
	}

	if db.MailHardDeleteByIds(idsSlice) {
		c.OK(c.Tr("common.success"))
	} else {
		c.Fail(-1, c.Tr("common.fail"))
	}
}

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
