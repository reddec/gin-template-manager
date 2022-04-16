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

// View uses Register to create template and creates Gin Handler.
// Data from callback (optional) will be passed as view param.
func (mgr *Manager) View(handler func(ctx context.Context) (interface{}, error), pageTemplate string) gin.HandlerFunc {
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
