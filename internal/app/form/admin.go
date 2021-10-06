package form

type AdminCreateUser struct {
	UserName string `binding:"Required;AlphaDashDot;MaxSize(35)"`
	Password string `binding:"MaxSize(255)"`
}
