package main

import (
	"embed"
	"flag"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/reddec/gin-template-manager"
)

//go:embed assets
var Assets embed.FS

func main() {
	prod := flag.Bool("prod", false, "enable production mode")
	bind := flag.String("bind", "127.0.0.1:8080", "bind address")
	flag.Parse()

	var templates *manager.Manager

	if *prod {
		gin.SetMode(gin.ReleaseMode)
		templates = manager.New(Assets, manager.Cache())
	} else {
		templates = manager.NewFromDir(".")
	}
	templates.Add("index", "assets/pages/index.html", "assets/layouts/base.html")

	router := gin.Default()
	router.HTMLRender = templates

	router.GET("/", func(gctx *gin.Context) {
		gctx.HTML(http.StatusOK, "index", "")
	})

	_ = router.Run(*bind)
}
