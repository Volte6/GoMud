package plugins

type WebConfig struct {
	navLinks map[string]string  // name=>path
	pages    map[string]WebPage // path=>WebPage
}

type WebPage struct {
	Name         string
	Path         string
	Filepath     string
	DataFunction func() map[string]any
}

func newWebConfig() WebConfig {
	return WebConfig{
		navLinks: map[string]string{},
		pages:    map[string]WebPage{},
	}
}

func (w *WebConfig) NavLink(name string, path string) {
	w.navLinks[name] = path
}

func (w *WebConfig) WebPage(name string, path string, file string, addToNav bool, dataFunc func() map[string]any) {
	if addToNav {
		w.NavLink(name, path)
	}
	w.pages[path] = WebPage{
		Name:         name,
		Path:         path,
		Filepath:     file,
		DataFunction: dataFunc,
	}
}
