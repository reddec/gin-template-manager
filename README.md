# Gin template manager

Manages Golang templates and integrates with Gin framework.

Features:

- Layouts, including nested ('base', 'sub-base', 'sub-sub-base', ...)
- Caching
- Streaming
- Supports embedded assets

Check [example](example) directory.

## Usage

**Initialize manager**

* Opt: filesystem: `templates := mananger.NewFromDir("path-to-dir")`
* Opt: assets (embedded): `templates := mananger.New(assets)`

**Register views**

* `templates.Register("index.html", "base.html", "child.html")`

**Link to Gin router**

* `router.HTMLRender = templates`

**Render**

* `gctx.HTML(http.StatusOK, "hello", "params")`

### Example

```go
templates := mananger.NewFromDir("path-to-dir")
templates.Register("hello.html", "base.html")
// in case your assets are NOT following convention you may use
// templates.Add("hello", "hello.html", "layouts/base.html")

router := gin.Default()
router.HTMLRender = templates

router.GET("/", func(gctx *gin.Context) {
    gctx.HTML(http.StatusOK, "hello.html", "params")
})
// ...
```

## Convention

Directory structure

* `pages` - contains views
* `layouts` - contains layouts for pages