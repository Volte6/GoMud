package web

import (
	"net/http"
	"text/template"

	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/mudlog"
)

func serveHome(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.New("index.html").Funcs(funcMap).ParseFiles(configs.GetConfig().FolderHtmlFiles.String()+"/public/_header.html", configs.GetConfig().FolderHtmlFiles.String()+"/public/index.html", configs.GetConfig().FolderHtmlFiles.String()+"/public/_footer.html")
	if err != nil {
		mudlog.Error("HTML ERROR", "error", err)
	}

	tmpl.Execute(w, configs.GetConfig())
	//GetStats()
}

func serveClient(w http.ResponseWriter, r *http.Request) {
	// read contents of webclient.html and print it out
	http.ServeFile(w, r, configs.GetConfig().FolderHtmlFiles.String()+"/public/webclient.html")
}
