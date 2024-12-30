package web

import (
	"log/slog"
	"net/http"
	"text/template"

	"github.com/volte6/gomud/internal/configs"
)

func adminIndex(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.New("index.html").Funcs(funcMap).ParseFiles(configs.GetConfig().FolderHtmlFiles.String()+"/admin/_header.html", configs.GetConfig().FolderHtmlFiles.String()+"/admin/index.html", configs.GetConfig().FolderHtmlFiles.String()+"/admin/_footer.html")
	if err != nil {
		slog.Error("HTML ERROR", "error", err)
	}

	tmpl.Execute(w, nil)

}
