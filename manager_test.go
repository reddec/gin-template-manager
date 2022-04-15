package manager_test

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	mananger "github.com/reddec/gin-template-manager"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_basic(t *testing.T) {
	mgr := mananger.NewFromDir("test-data")
	mgr.Add("hello", "hello.html", "layouts/base.html")

	tpl, err := mgr.Get("hello")
	require.NoError(t, err)
	var buffer bytes.Buffer
	err = tpl.Execute(&buffer, "xyz")
	require.NoError(t, err)

	text := buffer.String()

	assert.Contains(t, text, "<title>Hello world</title>")
	assert.Contains(t, text, "1234 xyz")
}

func TestManager_nestedLayouts(t *testing.T) {
	mgr := mananger.NewFromDir("test-data")
	mgr.Add("hello", "nested.html", "layouts/base.html", "layouts/child.html")

	tpl, err := mgr.Get("hello")
	require.NoError(t, err)
	var buffer bytes.Buffer
	err = tpl.Execute(&buffer, "xyz")
	require.NoError(t, err)

	text := buffer.String()

	assert.Contains(t, text, "<title>Hello world</title>")
	assert.Contains(t, text, "<main>main xyz</main>")
	assert.Contains(t, text, "<footer>footer xyz</footer>")
}

func TestManager_integration(t *testing.T) {
	templates := mananger.NewFromDir("test-data")
	templates.Add("hello", "hello.html", "layouts/base.html")

	router := gin.Default()
	router.HTMLRender = templates

	router.GET("/", func(gctx *gin.Context) {
		gctx.HTML(http.StatusOK, "hello", "qwerty")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "1234 qwerty")
}

func TestCache(t *testing.T) {
	workDir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(workDir)

	templateFile := filepath.Join(workDir, "index.html")
	err = ioutil.WriteFile(templateFile, []byte("hello world"), 0755)
	require.NoError(t, err)

	templates := mananger.NewFromDir(workDir, mananger.Cache())
	templates.Add("hello", "index.html")
	nonCached := mananger.NewFromDir(workDir)
	nonCached.Add("hello", "index.html")

	// works before
	tpl, err := templates.Get("hello")
	require.NoError(t, err)
	require.NotNil(t, tpl)

	tpl, err = nonCached.Get("hello")
	require.NoError(t, err)
	require.NotNil(t, tpl)

	// remove file, but cache should still work
	err = os.RemoveAll(templateFile)
	require.NoError(t, err)

	tpl, err = templates.Get("hello")
	require.NoError(t, err)
	require.NotNil(t, tpl)

	tpl, err = nonCached.Get("hello")
	require.Error(t, err)
	require.Nil(t, tpl)
}

func TestManager_Compile(t *testing.T) {
	workDir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(workDir)

	templateFile := filepath.Join(workDir, "index.html")
	err = ioutil.WriteFile(templateFile, []byte("hello world"), 0755)
	require.NoError(t, err)

	templates := mananger.NewFromDir(workDir, mananger.Cache())
	templates.Add("hello", "index.html")

	err = templates.Compile()
	require.NoError(t, err)

	err = os.RemoveAll(templateFile)
	require.NoError(t, err)

	tpl, err := templates.Get("hello")
	require.NoError(t, err)
	require.NotNil(t, tpl)
}

func TestWithoutStream(t *testing.T) {
	templates := mananger.NewFromDir("test-data",
		mananger.Func("doFunc", func() (string, error) {
			return "", fmt.Errorf("oops")
		}),
	)
	templates.Add("func", "func.html", "layouts/base.html")

	router := gin.Default()
	router.HTMLRender = templates

	router.GET("/", func(gctx *gin.Context) {
		gctx.HTML(http.StatusOK, "func", "qwerty")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestStream(t *testing.T) {
	templates := mananger.NewFromDir("test-data",
		mananger.Stream(),
		mananger.Func("doFunc", func() (string, error) {
			return "", fmt.Errorf("oops")
		}),
	)
	templates.Add("func", "func.html", "layouts/base.html")

	router := gin.Default()
	router.HTMLRender = templates

	router.GET("/", func(gctx *gin.Context) {
		gctx.HTML(http.StatusOK, "func", "qwerty")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code) // can not handle error
}

func TestFuncMap(t *testing.T) {
	mgr := mananger.NewFromDir("test-data", mananger.FuncMap(template.FuncMap{
		"doFunc": func() string { return "a1b2c3" },
	}))
	mgr.Add("func", "func.html", "layouts/base.html")

	tpl, err := mgr.Get("func")
	require.NoError(t, err)
	var buffer bytes.Buffer
	err = tpl.Execute(&buffer, "xyz")
	require.NoError(t, err)

	text := buffer.String()

	assert.Contains(t, text, "<title>a1b2c3</title>")
}
