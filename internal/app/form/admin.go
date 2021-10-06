package form

type AdminCreateUser struct {
	UserName string `binding:"Required;AlphaDashDot;MaxSize(35)"`
	Password string `binding:"MaxSize(255)"`
}

type AdminEditUser struct {
	LoginType        string `binding:"Required"`
	LoginName        string
	FullName         string `binding:"MaxSize(100)"`
	Email            string `binding:"Required;Email;MaxSize(254)"`
	Password         string `binding:"MaxSize(255)"`
	Website          string `binding:"MaxSize(50)"`
	Location         string `binding:"MaxSize(50)"`
	MaxRepoCreation  int
	Active           bool
	Admin            bool
	AllowGitHook     bool
	AllowImportLocal bool
	ProhibitLogin    bool
}
