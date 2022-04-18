package manager

import (
	"net/url"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

// View creates view context which could be used in template. Provides useful methods and exposes user data
// via .Data method.
func View(gctx *gin.Context, userData interface{}) *ViewContext {
	return &ViewContext{
		gctx:     gctx,
		rootPath: rootPath(gctx.Request.URL.Path),
		data:     userData,
	}
}

// ViewContext is convenient wrapper around context and data which could be used in templates.
type ViewContext struct {
	gctx     *gin.Context
	rootPath string
	data     interface{}
}

// Context used by handler.
func (vc *ViewContext) Context() *gin.Context {
	return vc.gctx
}

// Data provided by user
func (vc *ViewContext) Data() interface{} {
	return vc.data
}

// Rel constructs relative path from absolute.
func (vc *ViewContext) Rel(href string, paths ...string) string {
	href = path.Clean(path.Join(append([]string{href}, paths...)...))
	u, err := url.Parse(href)
	if err != nil || u.IsAbs() || !strings.HasPrefix(href, "/") {
		return href
	}

	return path.Clean("./" + vc.rootPath + href)
}

// RelArgs constructs relative path from absolute and povidided path parts.
func (vc *ViewContext) RelArgs(href string, paths []string) string {
	return vc.Rel(href, paths...)
}

// In checks is current path equal or under provided href.
func (vc *ViewContext) In(href string) bool {
	return href == vc.gctx.Request.URL.Path || strings.HasPrefix(vc.gctx.Request.URL.Path, path.Clean(href))
}

// Path from request.
func (vc *ViewContext) Path() string {
	return vc.gctx.Request.URL.Path
}

func rootPath(path string) string {
	if path == "" || !strings.HasPrefix(path, "/") {
		return path
	}
	segments := strings.Count(path, "/") - 1
	var res = make([]string, 0, segments)

	for i := 0; i < segments; i++ {
		res = append(res, "..")
	}
	return strings.Join(res, "/")
}
