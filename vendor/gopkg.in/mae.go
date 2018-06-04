package main

import (
	"github.com/go-macaron/gzip"
	"gopkg.in/macaron.v1"
)

func main() {
	m := macaron.New()
	m.Use(macaron.Logger())
	m.Use(macaron.Recovery())
	m.Use(gzip.Gziper())
	m.Use(macaron.Static("public"))
	// 注册路由
	m.Run()
}
