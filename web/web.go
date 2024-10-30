package web

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/volte6/gomud/configs"
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
	http.Handle("GET /static/public/", http.StripPrefix("/static/public/", http.FileServer(http.Dir("web/html/static/public"))))
	http.Handle("GET /static/admin/", doBasicAuth(handlerToHandlerFunc(http.StripPrefix("/static/admin/", http.FileServer(http.Dir("web/html/static/admin"))))))

	// Admin tools
	http.HandleFunc("GET /admin/", doBasicAuth(adminIndex))
	// Item Admin
	http.HandleFunc("GET /admin/items", doBasicAuth(itemsIndex))
	http.HandleFunc("GET /admin/items/itemdata", doBasicAuth(itemData))

	go func() {
		defer wg.Done()
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Error starting web server", "error", err)
		}
	}()

}

func serveHome(w http.ResponseWriter, r *http.Request) {

	stats := GetStats()

	strB := strings.Builder{}

	strB.WriteString("<html><head><title>GoMud Configuration</title><style>\n")

	strB.WriteString(
		"body {\n" +
			"font-family: Verdana, sans-serif;\n" +
			"}\n" +
			"th {\n" +
			"background-color:#ccc;" +
			"}\n" +
			"tr {\n" +
			"\tborder-bottom: 1px solid #ddd;\n" +
			"}\n" +
			"tr:nth-child(even) { \n" +
			"\tbackground-color: #D6EEEE;\n" +
			"}\n" +
			"td {\n" +
			"font-family: monospace;\n" +
			"}\n" +
			".footer{\n" +
			"text-align:center;\n" +
			"}\n")

	strB.WriteString("</style></head><body>\n")
	strB.WriteString("<h1>GoMud</h1>\n")

	if stats.WebSocketPort > 0 {
		strB.WriteString("<p><b>Access Web Terminal:</b> <a href=\"/webclient\">Link</a></p>\n")
	}

	strB.WriteString("<p>&nbsp;</p>\n")

	strB.WriteString("<h3>Players Online: </h1>\n")

	if len(stats.OnlineUsers) > 0 {
		strB.WriteString("<table border=\"1\" cellspacing=\"0\" cellpadding=\"3\">\n")
		strB.WriteString("<tr><th>#</th><th>Character</th><th>Level</th><th>Alignment</th><th>Profession</th><th>Time Online</th><th>Permission</th></tr>\n")
		for i, oInfo := range stats.OnlineUsers {

			if oInfo.IsAFK {
				oInfo.OnlineTimeStr += ` (AFK)`
			}

			strB.WriteString(fmt.Sprintf(`<tr><td align="right">%d.</td><td align="center"><b>%s</b></td><td align="center">%d</td><td align="center">%s</td><td align="center">%s</td><td align="center">%s</td><td align="center">%s</td></tr>`+"\n",
				i+1,
				oInfo.CharacterName,
				oInfo.Level,
				oInfo.Alignment,
				oInfo.Profession,
				oInfo.OnlineTimeStr,
				oInfo.Permission,
			))
		}
		strB.WriteString("</table>\n")
	} else {
		strB.WriteString(`<i>None</i>`)
	}

	strB.WriteString("<p>&nbsp;</p>\n")

	strB.WriteString("<h3>Server Config: </h1>\n")

	// exclude port, seed, and filepath info from webpage
	allConfigData := configs.GetConfig().AllConfigData(`*port`, `seed`, `folder*`, `file*`)

	// Extract keys into a slice
	keys := make([]string, 0, len(allConfigData))
	for key := range allConfigData {
		keys = append(keys, key)
	}

	// Sort the keys
	sort.Strings(keys)

	strB.WriteString("<table border=\"1\" cellspacing=\"0\" cellpadding=\"3\">\n")
	strB.WriteString("<tr><th>Name</th><th>Value</th></tr>\n")
	for _, k := range keys {
		displayName := strings.Replace(k, ` (locked)`, ``, -1)
		strB.WriteString(fmt.Sprintf("<tr><td><b>%s</b></td><td>%v</td></tr>\n", displayName, allConfigData[k]))
	}
	strB.WriteString("</table>\n")

	strB.WriteString("<p class=\"footer\">Powered by <b>GoMud</b> - Available free at <a href=\"http://github.com/Volte6/GoMud\">github.com/Volte6/GoMud</a></p>")
	strB.WriteString("</body></html>")

	w.Write([]byte(strB.String()))
}

func serveClient(w http.ResponseWriter, r *http.Request) {
	// read contents of webclient.html and print it out
	http.ServeFile(w, r, "web/html/public/webclient.html")
}

func Shutdown() {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown failed:%+v", err)
	}

}
