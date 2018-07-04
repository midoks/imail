package app

import (
	""
)

const (
	HOME = "home"
)

type Context struct {
	*macaron.Context
	IsLogged bool
}

// func Home(ctx *macaron.Context) {
// 	ctx.Data["Name"] = "jeremy"
// 	ctx.HTML(200, HOME) // 200 is the response code.
// }

func Home(ctx *Context) {
	ctx.Data["Name"] = "jeremy"
	ctx.HTML(200, HOME) // 200 is the response code.
}
