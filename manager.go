package manager

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin/render"
)

// New template manager which supports layouts and can be used as template engine for Gin.
func New(fs fs.FS, options ...Option) *Manager {
	mgr := &Manager{
		fs:        fs,
		templates: make(map[string][]string),
		funcMap:   make(template.FuncMap),
	}
	for _, opt := range options {
		opt(mgr)
	}
	return mgr
}

// NewFromDir creates new template manager based on filesystem directory.
func NewFromDir(dir string, options ...Option) *Manager {
	return New(os.DirFS(dir), options...)
}

// Manager of templates. Supports caching (builds template once), streaming (render directly to client),
// and functions map (ex: Sprig).
type Manager struct {
	fs        fs.FS
	templates map[string][]string
	stream    bool
	funcMap   template.FuncMap
	cache     struct {
		enable    bool
		templates sync.Map // map[string]*cachedTemplate
	}
}

// Add template as from template file (related to FS used during creation) and optional layouts.
// Configuration will be saved under provided name. Template will be lazy-compiled. See Compile
// for earlier warm-up. Thread UNSAFE, since templates are usually supposed to be added once during the initialization
// phase.
func (mgr *Manager) Add(name string, templateFile string, layouts ...string) {
	mgr.templates[name] = append(layouts, templateFile)
}

// Get and parse template (optionally from cache) by the same name as used in Add.
// If cache enabled, template will be compiled only once. Thread-safe.
func (mgr *Manager) Get(name string) (*template.Template, error) {
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
	for name := range mgr.templates {
		if _, err := mgr.Get(name); err != nil {
			return fmt.Errorf("compile %s: %w", name, err)
		}
	}
	return nil
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
	files := mgr.templates[name]
	templ := template.New("").Funcs(mgr.funcMap)

	for _, file := range files {
		content, err := fs.ReadFile(mgr.fs, file)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", file, err)
		}
		t, err := templ.Parse(string(content))
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", file, err)
		}
		templ = t
	}
	return templ, nil
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
