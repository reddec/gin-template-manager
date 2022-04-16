package manager

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path"
	"sync"

	"github.com/gin-gonic/gin/render"
)

const LayoutFile = "@layout.html"

// New template manager which supports layouts and can be used as template engine for Gin.
func New(fs fs.FS, options ...Option) *Manager {
	mgr := &Manager{
		fs:      fs,
		alias:   make(map[string]string),
		funcMap: make(template.FuncMap),
	}
	for _, opt := range options {
		opt(mgr)
	}
	return mgr
}

// Manager of templates. Supports caching (builds template once), streaming (render directly to client),
// and functions map (ex: Sprig).
type Manager struct {
	fs      fs.FS
	alias   map[string]string
	stream  bool
	funcMap template.FuncMap
	cache   struct {
		enable    bool
		templates sync.Map // map[string]*cachedTemplate
	}
}

// Alias template file (related to FS used during creation) and to short name.
// Thread UNSAFE, since templates are usually supposed to be added once during the initialization
// phase.
func (mgr *Manager) Alias(aliasName, templateFile string) {
	mgr.alias[aliasName] = templateFile
}

// Get and parse template (optionally from cache) by the template name or alias.
// If cache enabled, template will be compiled only once. Thread-safe.
func (mgr *Manager) Get(name string) (*template.Template, error) {
	if realName, ok := mgr.alias[name]; ok {
		name = realName
	}
	if !mgr.cache.enable {
		return mgr.compile(name)
	}

	cachedRawView, _ := mgr.cache.templates.LoadOrStore(name, &cachedTemplate{})
	cachedView := cachedRawView.(*cachedTemplate)

	cachedView.lock.RLock()
	view := cachedView.view
	cachedView.lock.RUnlock()
	if view != nil {
		return view, nil
	}

	cachedView.lock.Lock()
	defer cachedView.lock.Unlock()
	if cachedView.view != nil {
		return cachedView.view, nil
	}

	t, err := mgr.compile(name)
	if err != nil {
		return nil, err
	}
	cachedView.view = t

	return t, nil
}

// Compile all templates. There is no sense to use it without enabled caching. Could be used as warm-up.
func (mgr *Manager) Compile() error {
	return fs.WalkDir(mgr.fs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() == LayoutFile {
			return nil
		}
		_, err = mgr.Get(path)
		return err
	})
}

// Instance of Gin renderer, used by Gin itself.
func (mgr *Manager) Instance(name string, params interface{}) render.Render {
	view, err := mgr.Get(name)
	if err != nil {
		return &failedRender{err: err, templateName: name}
	}
	return &renderer{
		view:   view,
		params: params,
		stream: mgr.stream,
	}
}

func (mgr *Manager) compile(name string) (*template.Template, error) {
	root, err := mgr.layout(path.Dir(name), template.New("").Funcs(mgr.funcMap))
	if err != nil {
		return nil, fmt.Errorf("layouts form %s: %w", name, err)
	}

	content, err := fs.ReadFile(mgr.fs, name)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", name, err)
	}
	return root.Parse(string(content))
}

func (mgr *Manager) layout(name string, root *template.Template) (*template.Template, error) {
	parent := path.Dir(name)
	if !(name == "" || name == "/" || name == ".") {
		top, err := mgr.layout(parent, root)
		if err != nil {
			return nil, fmt.Errorf("get parent layout %s: %w", parent, err)
		}
		root = top
	}

	layoutFile := path.Join(name, LayoutFile)
	content, err := fs.ReadFile(mgr.fs, layoutFile)
	if errors.Is(err, fs.ErrNotExist) {
		return root, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read layout %s: %w", layoutFile, err)
	}
	t, err := root.Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", layoutFile, err)
	}

	return t, nil
}

type renderer struct {
	view   *template.Template
	params interface{}
	stream bool
}

func (tr *renderer) Render(writer http.ResponseWriter) error {
	if tr.stream {
		return tr.view.Execute(writer, tr.params)
	}
	var buffer bytes.Buffer
	if err := tr.view.Execute(&buffer, tr.params); err != nil {
		return err
	}
	_, err := writer.Write(buffer.Bytes())
	return err
}

func (tr *renderer) WriteContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html")
}

type failedRender struct {
	err          error
	templateName string
}

func (fr *failedRender) Render(writer http.ResponseWriter) error {
	return fmt.Errorf("render %s: %w", fr.templateName, fr.err)
}

func (fr *failedRender) WriteContentType(w http.ResponseWriter) {}

type cachedTemplate struct {
	lock sync.RWMutex
	view *template.Template
}
