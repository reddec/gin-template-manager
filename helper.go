package manager

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	pagesDir   = "pages"
	layoutsDir = "layouts"
)

// Register template record based on convention: pages are stored in `pages` directory and layouts under `layouts`.
func (mgr *Manager) Register(pageTemplate string, layouts ...string) {
	var layoutNames = make([]string, 0, len(layouts))
	for _, l := range layouts {
		layoutNames = append(layoutNames, layoutsDir+"/"+l)
	}
	mgr.Add(pageTemplate, pagesDir+"/"+pageTemplate, layoutNames...)
}

// View uses Register to create template and creates Gin Handler.
// Data from callback (optional) will be passed as view param.
func (mgr *Manager) View(handler func(ctx context.Context) (interface{}, error), pageTemplate string, layouts ...string) gin.HandlerFunc {
	mgr.Register(pageTemplate, layouts...)
	return func(gctx *gin.Context) {
		var params interface{}
		if handler != nil {
			data, err := handler(gctx.Request.Context())
			if err != nil {
				_ = gctx.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			params = data
		}
		gctx.HTML(http.StatusOK, pageTemplate, params)
	}
}
