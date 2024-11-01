package web

import (
	"log/slog"
	"net/http"
	"text/template"
)

func adminIndex(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.New("index.html").Funcs(funcMap).ParseFiles("_datafiles/html/admin/_header.html", "_datafiles/html/admin/index.html", "_datafiles/html/admin/_footer.html")
	if err != nil {
		slog.Error("HTML ERROR", "error", err)
	}

	tmpl.Execute(w, nil)

}
