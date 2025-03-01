package web

import (
	"net/http"
	"text/template"

	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/mudlog"
)

func serveOnline(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.New("online.html").Funcs(funcMap).ParseFiles(configs.GetConfig().FolderHtmlFiles.String()+"/public/_header.html", configs.GetConfig().FolderHtmlFiles.String()+"/public/online.html", configs.GetConfig().FolderHtmlFiles.String()+"/public/_footer.html")
	if err != nil {
		mudlog.Error("HTML ERROR", "error", err)
	}

	tmpl.Execute(w, GetStats())
}
