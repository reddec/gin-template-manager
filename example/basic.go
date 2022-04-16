package main

import (
	"embed"
	"flag"
	"io/fs"
	"net/http"
	"os"
	"path"

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
		gctx.HTML(http.StatusOK, "index.html", manager.View(gctx, nil))
	})

	router.GET("/hello", func(gctx *gin.Context) {
		gctx.HTML(http.StatusOK, "hello.html", manager.View(gctx, nil))
	})
	router.GET("/features", func(gctx *gin.Context) {
		gctx.HTML(http.StatusOK, "features/index.html", manager.View(gctx, nil))
	})
	router.GET("/features/:feature", func(gctx *gin.Context) {
		feature := path.Base(path.Clean(gctx.Param("feature")))
		gctx.HTML(http.StatusOK, "features/"+feature+".html", manager.View(gctx, nil))
	})

	_ = router.Run(*bind)
}
