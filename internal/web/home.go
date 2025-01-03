package web

import (
	"log/slog"
	"net/http"
	"text/template"

	"github.com/volte6/gomud/internal/configs"
)

func serveHome(w http.ResponseWriter, r *http.Request) {

	homepageData := struct {
		Stats      Stats
		ConfigData map[string]any
	}{
		GetStats(),
		configs.GetConfig().AllConfigData(`*port`, `seed`, `folder*`, `file*`),
	}

	tmpl, err := template.New("index.html").Funcs(funcMap).ParseFiles(configs.GetConfig().FolderHtmlFiles.String() + "/public/index.html")
	if err != nil {
		slog.Error("HTML ERROR", "error", err)
	}

	tmpl.Execute(w, homepageData)

}

func serveClient(w http.ResponseWriter, r *http.Request) {
	// read contents of webclient.html and print it out
	http.ServeFile(w, r, configs.GetConfig().FolderHtmlFiles.String()+"/public/webclient.html")
}
