package form

type AdminCreateDomain struct {
	Domain string `binding:"Required;AlphaDashDot;MaxSize(255)"`
}

type AdminDeleteDomain struct {
	Id int64 `binding:"Required;AlphaDashDot;MaxSize(255)"`
}
