package form

type AdminCreateUser struct {
	LoginType  string `binding:"Required"`
	LoginName  string
	UserName   string `binding:"Required;AlphaDashDot;MaxSize(35)"`
	Password   string `binding:"MaxSize(255)"`
	SendNotify bool
}
