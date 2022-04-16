package manager

import "html/template"

type Option func(mgr *Manager)

// Stream response directly to client. Reduces memory consumption, but it will not let system send valid 5xx code
// in case of rendering failure.
func Stream() Option {
	return func(mgr *Manager) {
		mgr.stream = true
	}
}

// Cache compiled template. Slightly increases memory usage. Recommended in production environment.
func Cache() Option {
	return func(mgr *Manager) {
		mgr.cache.enable = true
	}
}

// FuncMap sets user-defined functions for templates. For example: Sprig library.
func FuncMap(functions template.FuncMap) Option {
	return func(mgr *Manager) {
		mgr.funcMap = functions
	}
}

// Func adds templates functions to internal map.
func Func(name string, fn interface{}) Option {
	return func(mgr *Manager) {
		mgr.funcMap[name] = fn
	}
}
