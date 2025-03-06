package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"
	"time"

	"github.com/gorilla/websocket"
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/util"
)

var (
	httpServer *http.Server

	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	httpRoot string
)

// serveTemplate searches for the requested file in the HTTP_ROOT,
// parses it as a template, and serves it.
func serveTemplate(w http.ResponseWriter, r *http.Request) {

	if httpRoot == "" {
		httpRoot = configs.GetConfig().FolderPublicHtml.String()
	}

	// Clean the path to prevent directory traversal.
	reqPath := filepath.Clean(r.URL.Path) // Example: / or /info/faq

	// Build the full file path.
	fullPath := filepath.Join(httpRoot, reqPath)

	// If the path is a directory, look for an index.html.
	info, err := os.Stat(fullPath)
	if err != nil {
		if filepath.Ext(fullPath) != ".html" {
			fullPath += ".html"
		}
	} else if info.IsDir() {
		fullPath = filepath.Join(fullPath, "index.html")
	}

	fileExt := filepath.Ext(fullPath)
	fileBase := filepath.Base(fullPath)

	// Check if the file exists, else 404
	fInfo, err := os.Stat(fullPath)
	if err != nil || len(fileBase) > 0 && fileBase[0] == '_' {
		mudlog.Info("HTTP", "ip", r.RemoteAddr, "ref", r.Header.Get("Referer"), "filePath", fullPath, "fileExtension", fileExt, "error", "Not found")

		fullPath = filepath.Join(httpRoot, `404.html`)
		fInfo, err = os.Stat(fullPath)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}

	// Log the request
	mudlog.Info("HTTP", "ip", r.RemoteAddr, "ref", r.Header.Get("Referer"), "filePath", fullPath, "fileExtension", fileExt, "size", fmt.Sprintf(`%.2fk`, float64(fInfo.Size())/1024))

	// For non-HTML files, serve them statically.
	if fileExt != ".html" {
		http.ServeFile(w, r, fullPath)
		return
	}

	templateData := map[string]interface{}{
		"REQUEST": r,
		"CONFIG":  configs.GetConfig(),
		"STATS":   GetStats(),
	}

	templateFiles := []string{}

	if _, err := os.Stat(filepath.Join(filepath.Dir(fullPath), `_header.html`)); err == nil {
		templateFiles = append(templateFiles, filepath.Join(filepath.Dir(fullPath), `_header.html`))
	} else if _, err := os.Stat(filepath.Join(httpRoot, `_header.html`)); err == nil {
		templateFiles = append(templateFiles, filepath.Join(httpRoot, `_header.html`))
	}

	if _, err := os.Stat(filepath.Join(filepath.Dir(fullPath), `_footer.html`)); err == nil {
		templateFiles = append(templateFiles, filepath.Join(filepath.Dir(fullPath), `_footer.html`))
	} else if _, err := os.Stat(filepath.Join(httpRoot, `_footer.html`)); err == nil {
		templateFiles = append(templateFiles, filepath.Join(httpRoot, `_footer.html`))
	}

	// Add the final (actual) file
	templateFiles = append(templateFiles, fullPath)

	tmpl, err := template.New(filepath.Base(fullPath)).Funcs(funcMap).ParseFiles(templateFiles...)
	if err != nil {
		mudlog.Error("HTML ERROR", "action", "ParseFiles", "error", err)
		http.Error(w, "Error executing template", http.StatusInternalServerError)
	}

	// Execute the template and write it to the response.
	if err := tmpl.Execute(w, templateData); err != nil {
		mudlog.Error("HTML ERROR", "action", "Execute", "error", err)
		http.Error(w, "Error executing template", http.StatusInternalServerError)
	}
}

func Listen(webPort int, wg *sync.WaitGroup, webSocketHandler func(*websocket.Conn)) {

	mudlog.Info("Starting web server", "webport", webPort)

	wg.Add(1)

	// HTTP Server
	httpServer = &http.Server{Addr: fmt.Sprintf(`:%d`, webPort)}

	// Routing
	// Basic homepage
	http.HandleFunc("/", serveTemplate)

	// websocket upgrade
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("WebSocket upgrade failed:", err)
			return
		}
		defer conn.Close()

		webSocketHandler(conn)
	})

	http.Handle("GET /admin/static/", RunWithMUDLocked(
		doBasicAuth(
			handlerToHandlerFunc(
				http.StripPrefix("/admin/static/", http.FileServer(http.Dir(configs.GetConfig().FolderAdminHtml.String()+"/static"))),
			),
		),
	))

	// Admin tools
	http.HandleFunc("GET /admin/", RunWithMUDLocked(
		doBasicAuth(adminIndex),
	))

	// Item Admin
	http.HandleFunc("GET /admin/items/", RunWithMUDLocked(
		doBasicAuth(itemsIndex),
	))
	http.HandleFunc("GET /admin/items/itemdata/", RunWithMUDLocked(
		doBasicAuth(itemData),
	))

	// Race Admin
	http.HandleFunc("GET /admin/races/", RunWithMUDLocked(
		doBasicAuth(racesIndex)),
	)
	http.HandleFunc("GET /admin/races/racedata/", RunWithMUDLocked(
		doBasicAuth(raceData)),
	)

	// Mob Admin
	http.HandleFunc("GET /admin/mobs/", RunWithMUDLocked(
		doBasicAuth(mobsIndex),
	))
	http.HandleFunc("GET /admin/mobs/mobdata/", RunWithMUDLocked(
		doBasicAuth(mobData),
	))

	// Mutator Admin
	http.HandleFunc("GET /admin/mutators/", RunWithMUDLocked(
		doBasicAuth(mutatorsIndex),
	))
	http.HandleFunc("GET /admin/mutators/mutatordata/", RunWithMUDLocked(
		doBasicAuth(mutatorData),
	))

	// Room Admin
	http.HandleFunc("GET /admin/rooms/", RunWithMUDLocked(
		doBasicAuth(roomsIndex),
	))
	http.HandleFunc("GET /admin/rooms/roomdata/", RunWithMUDLocked(
		doBasicAuth(roomData),
	))

	go func() {
		defer wg.Done()
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			mudlog.Error("Error starting web server", "error", err)
		}
	}()

}

// This wraps the handler functiojn with a game lock (mutex) to keep the mud from
// Concurrently accessing the same memory
func RunWithMUDLocked(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		util.LockMud()
		defer util.UnlockMud()

		next.ServeHTTP(w, r)
	})
}

func Shutdown() {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown failed:%+v", err)
	}

}

func DataFiles() string {
	return configs.GetConfig().FolderDataFiles.String()
}

func sendError(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	if status == http.StatusNotFound {
		fmt.Fprint(w, "custom 404")
	}
}
