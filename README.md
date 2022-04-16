# Gin template manager

Manages Golang templates and integrates with Gin framework.

Features:

- Layouts, including nested ('base', 'sub-base', 'sub-sub-base', ...)
- Caching
- Streaming
- Supports embedded assets

Check [example](example) directory.

**STATUS:** proof-of-concept

## Usage

**Initialize manager**

* Opt: filesystem: `templates := mananger.New(os.DirFS("path-to-dir"))`
* Opt: assets (embedded): `templates := mananger.New(assets)`

**Link to Gin router**

* `router.HTMLRender = templates`

**Render**

* `gctx.HTML(http.StatusOK, "index.html", "params")`

### Example

```go
templates := mananger.New(os.DirFS("path-to-dir"))

router := gin.Default()
router.HTMLRender = templates

router.GET("/", func(gctx *gin.Context) {
    gctx.HTML(http.StatusOK, "index.html", "params")
})
// ...
```

## Convention

Directory structure

* `@layout.html` - layout file