package webclient

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
	"github.com/volte6/mud/configs"
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
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/client", serveClient)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("WebSocket upgrade failed:", err)
			return
		}
		defer conn.Close()

		webSocketHandler(conn)
	})

	go func() {
		defer wg.Done()
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Error starting web server", "error", err)
		}
	}()

}

func serveHome(w http.ResponseWriter, r *http.Request) {

	strB := strings.Builder{}

	strB.WriteString("<html><head><style>\n")

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

	allConfigData := configs.GetConfig().AllConfigData()

	// Extract keys into a slice
	keys := make([]string, 0, len(allConfigData))
	for key := range allConfigData {
		keys = append(keys, key)
	}

	// Sort the keys
	sort.Strings(keys)

	strB.WriteString("<p><a href=\"/client\">Log in using the web-based \"virtual terminal\"</a></p>\n")

	strB.WriteString("<table border=\"1\" cellspacing=\"0\" cellpadding=\"3\">\n")
	strB.WriteString("<tr><th>Name</th><th>Value</th></tr>\n")
	for _, k := range keys {
		displayName := strings.Replace(k, ` (locked)`, ``, -1)
		if strings.HasSuffix(strings.ToLower(displayName), `port`) || strings.ToLower(displayName) == "seed" {
			continue
		} else {
			strB.WriteString(fmt.Sprintf("<tr><td><b>%s</b></td><td>%v</td></tr>\n", displayName, allConfigData[k]))
		}
	}
	strB.WriteString("</table>\n")

	strB.WriteString("<p class=\"footer\">Powered by <b>GoMud</b> - Available free at <a href=\"http://github.com/Volte6/GoMud\">github.com/Volte6/GoMud</a></p>")
	strB.WriteString("</body></html>")

	w.Write([]byte(strB.String()))
}

func serveClient(w http.ResponseWriter, r *http.Request) {
	// read contents of webclient.html and print it out
	http.ServeFile(w, r, "webclient/webclient.html")
}

func Shutdown() {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown failed:%+v", err)
	}

}
