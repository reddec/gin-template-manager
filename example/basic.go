package main

import (
	"embed"
	"flag"
	"io/fs"
	"net/http"
	"os"

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
		root, _ := fs.Sub(&Assets, "assets")
		templates = manager.New(root, manager.Cache())
	} else {
		templates = manager.New(os.DirFS("assets"))
	}

	router := gin.Default()
	router.HTMLRender = templates

	router.GET("/", func(gctx *gin.Context) {
		gctx.HTML(http.StatusOK, "index.html", "")
	})
	// alternative
	router.GET("/hello", templates.View(nil, "hello.html"))

	_ = router.Run(*bind)
}
