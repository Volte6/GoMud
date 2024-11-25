package web

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/volte6/gomud/internal/util"
)

var (
	httpServer *http.Server

	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func Listen(webPort int, wg *sync.WaitGroup, webSocketHandler func(*websocket.Conn)) {

	slog.Info("Starting web server", "webport", webPort)

	wg.Add(1)

	// HTTP Server
	httpServer = &http.Server{Addr: fmt.Sprintf(`:%d`, webPort)}
	// Routing
	// Basic homepage
	http.HandleFunc("/", serveHome)
	// websocket client
	http.HandleFunc("/webclient", serveClient)
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

	// Static resources
	http.Handle("GET /static/public/", handlerToHandlerFunc(
		http.StripPrefix("/static/public/", http.FileServer(http.Dir("_datafiles/html/static/public"))),
	))

	http.Handle("GET /static/admin/", RunWithMUDLocked(
		doBasicAuth(
			handlerToHandlerFunc(
				http.StripPrefix("/static/admin/", http.FileServer(http.Dir("_datafiles/html/static/admin"))),
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

	go func() {
		defer wg.Done()
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Error starting web server", "error", err)
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
