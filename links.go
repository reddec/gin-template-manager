package manager

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/reddec/gin-template-manager/internal"
)

const linksKey = "links"

// NewLinks creates and installs alias handler.
func NewLinks(engine *gin.Engine) *Links {
	links := &Links{
		routes: make(map[string]internal.Path),
	}
	engine.Use(func(gctx *gin.Context) {
		gctx.Set(linksKey, links)
		gctx.Next()
	})

	return links
}

type Links struct {
	routes map[string]internal.Path
}

// Named registers route under provided name, which could be used in Views by .Ref function.
// Example:
//
//     gin.GET(links.Named("eventByID", "/event/:id"), func(gctx *gin.Context) {
//     })
//     ...
//
// Then in view we can use
//
//     {{.Ref "eventByID" 1234}}
//
// Which will generate relative URL to named route with proper parameters
func (mgr *Links) Named(name string, path string) string {
	mgr.routes[name] = internal.Path(path)
	return path
}

// Path to alias with arguments (escaped).
func (mgr *Links) Path(alias string, args ...interface{}) (string, error) {
	p, ok := mgr.routes[alias]
	if !ok {
		return "", fmt.Errorf("alias %s not found", alias)
	}

	var sArgs = make([]string, 0, len(args))
	for _, a := range args {
		sArgs = append(sArgs, fmt.Sprint(a))
	}
	return p.Build(sArgs...)
}
