package web

import (
	"net/http"
	"text/template"

	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/mudlog"
)

func viewConfig(w http.ResponseWriter, r *http.Request) {

	configData := configs.GetConfig().AllConfigData(`*port`, `seed*`, `folder*`, `file*`, `seedint`)

	tmpl, err := template.New("viewconfig.html").Funcs(funcMap).ParseFiles(configs.GetConfig().FolderHtmlFiles.String()+"/public/_header.html", configs.GetConfig().FolderHtmlFiles.String()+"/public/viewconfig.html", configs.GetConfig().FolderHtmlFiles.String()+"/public/_footer.html")
	if err != nil {
		mudlog.Error("HTML ERROR", "error", err)
	}

	tmpl.Execute(w, configData)

}
