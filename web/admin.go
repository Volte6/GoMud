package web

import (
	"net/http"
)

func serveAdmin(w http.ResponseWriter, r *http.Request) {

	p := r.PathValue("page")
	if p == `` {
		http.ServeFile(w, r, "web/html/admin/index.html")
		return
	}
	http.ServeFile(w, r, "web/html/admin/"+p+".html")

	// read contents of webclient.html and print it out
	// http.ServeFile(w, r, "webclient/webclient.html")
}

func submitAdminRequest(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("Accepting Post Data"))

	// read contents of webclient.html and print it out
	// http.ServeFile(w, r, "webclient/webclient.html")
}
