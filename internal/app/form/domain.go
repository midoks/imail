package form

type AdminCreateDomain struct {
	Domain string `binding:"Required;AlphaDashDot;MaxSize(255)"`
}
