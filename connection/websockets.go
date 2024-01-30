package connection

/*
import (
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type WebsocketConnection struct {
	conn websocket.Conn
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		//origin := r.Header.Get("Origin")
		//return origin == "http://localhost:8000"
		return true
	},
}

func (w *WebsocketConnection) Read(b []byte) (n int, err error) {
	_, p, err := w.conn.ReadMessage()
	if err != nil {
		return 0, err
	}
	copy(b, p)
	return len(p), nil
}

func (w *WebsocketConnection) Write(b []byte) (n int, err error) {
	if err := w.conn.WriteMessage(websocket.TextMessage, b); err != nil {
		return 0, err
	}
	return len(b), nil
}

func (w *WebsocketConnection) Close() error {
	return w.conn.Close()
}

func (w *WebsocketConnection) LocalAddr() net.Addr {
	return w.conn.LocalAddr()
}

func (w *WebsocketConnection) RemoteAddr() net.Addr {
	return w.conn.RemoteAddr()
}

func (w *WebsocketConnection) SetDeadline(t time.Time) error {
	return errors.New("not implemented for websocket")
}

func (w *WebsocketConnection) SetReadDeadline(t time.Time) error {
	return errors.New("not implemented for websocket")
}

func (w *WebsocketConnection) SetWriteDeadline(t time.Time) error {
	return errors.New("not implemented for websocket")
}

// Accepts an incoming connection and upgrades it to a WebSocket connection.
func handleWebConnection(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection to WebSocket", http.StatusBadRequest)
		return
	}


	//
	// Do something with the connection
	//
	for {
		messageType, p, err := conn.ReadMessage()

		// Handle any disconnections
		if err != nil {
			// Handle client disconnect here
			s.log.Infof("Client disconnected %s", conn.RemoteAddr())
			// Send a nil message to the response dispatcher to shut it down
			backToClient <- nil
			break
		}

		// Immediately ignore non-text messages
		if messageType != websocket.TextMessage {
			s.log.Errorf("Received non-text message type: %d", messageType)
			continue
		}

		s.log.Infof("Message (type %d) received: \t%s", messageType, p)

		go s.requestHandler(p, backToClient)

	}

}
*/
