package web

import (
	"net/http"
	"text/template"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

func adminIndex(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.New("index.html").Funcs(funcMap).ParseFiles(configs.GetFilePathsConfig().AdminHtml.String()+"/_header.html", configs.GetFilePathsConfig().AdminHtml.String()+"/index.html", configs.GetFilePathsConfig().AdminHtml.String()+"/_footer.html")
	if err != nil {
		mudlog.Error("HTML ERROR", "error", err)
	}

	tmpl.Execute(w, nil)

}
