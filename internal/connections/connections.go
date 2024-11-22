package connections

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/util"
	"gopkg.in/yaml.v2"
)

const ReadBufferSize = 1024

type ConnectionId = uint64

var (

	//
	// Mutex
	//
	lock sync.RWMutex = sync.RWMutex{}
	//
	// Counters
	//
	connectCounter    uint64 = 0 // a counter for each time a connection is accepted
	disconnectCounter uint64 = 0 // a counter for each tim ea connection is dropped
	//
	// Track connections
	//
	netConnections map[ConnectionId]*ConnectionDetails = map[ConnectionId]*ConnectionDetails{} // a mapping of unique id's to connections
	//
	// Channel to send a shutdown signal to
	//
	shutdownChannel chan os.Signal // channel to receive shutdown signals
)

func SignalShutdown(s os.Signal) {
	if shutdownChannel != nil {
		shutdownChannel <- s
	}
}

func Add(conn net.Conn, wsConn *websocket.Conn) *ConnectionDetails {

	lock.Lock()
	defer lock.Unlock()

	connectCounter++

	connDetails := NewConnectionDetails(
		connectCounter,
		conn,
		wsConn,
	)

	netConnections[connDetails.ConnectionId()] = connDetails

	// Temp hackish thing to send all color aliases
	if wsConn != nil {

		data := make(map[string]map[string]string, 100)
		if yfile, err := os.ReadFile(util.FilePath(string(configs.GetConfig().FileAnsiAliases))); err == nil {
			if err := yaml.Unmarshal(yfile, &data); err == nil {

				jsonString, _ := json.Marshal(data[`color256`])

				events.AddToQueue(events.WebClientCommand{
					ConnectionId: connDetails.ConnectionId(),
					Text:         "COLORMAP:" + string(jsonString) + "\n",
				})

			}
		}

	}

	// return the unique ID to find this connection later
	return connDetails
}

// Returns the total number of connections
func Get(id ConnectionId) *ConnectionDetails {
	lock.Lock()
	defer lock.Unlock()

	return netConnections[id]
}

func IsWebsocket(id ConnectionId) bool {
	lock.Lock()
	defer lock.Unlock()

	if cd, ok := netConnections[id]; ok {
		return cd.IsWebsocket()
	}

	return false
}

func GetAllConnectionIds() []ConnectionId {

	lock.Lock()
	defer lock.Unlock()

	ids := make([]ConnectionId, len(netConnections))

	for id := range netConnections {
		ids = append(ids, id)
	}

	return ids
}

func Cleanup() {
	for _, id := range GetAllConnectionIds() {
		Remove(id)
	}
}

func Kick(id ConnectionId) (err error) {

	lock.Lock()
	defer lock.Unlock()

	// Try to retrieve the value
	if cd, ok := netConnections[id]; ok {

		// close the connection, no longer useful.
		cd.Close()
		// keep track of the number of disconnects
		disconnectCounter++
		// remove the connection from the map
		slog.Info("connection kicked", "connectionId", id, "remoteAddr", cd.RemoteAddr().String())

		return nil

	}

	return errors.New("connection not found")
}

func Remove(id ConnectionId) (err error) {

	lock.Lock()
	defer lock.Unlock()

	// Try to retrieve the value
	if cd, ok := netConnections[id]; ok {

		// close the connection, no longer useful.
		cd.Close()
		// keep track of the number of disconnects
		disconnectCounter++
		// Remove the entry
		delete(netConnections, id)
		// remove the connection from the map
		slog.Info("connection removed", "connectionId", id, "remoteAddr", cd.RemoteAddr().String())

		return nil

	}

	return errors.New("connection not found")
}

func Broadcast(colorizedText []byte, rawText []byte) {

	lock.Lock()

	removeIds := []ConnectionId{}

	for id, cd := range netConnections {

		if cd.state == Login {
			continue
		}

		// Write the message to the connection
		var err error

		if cd.IsWebsocket() {
			_, err = cd.Write(rawText)
		} else {
			_, err = cd.Write(colorizedText)
		}

		if err != nil {
			slog.Error("could not write to connection", "connectionId", id, "remoteAddr", cd.RemoteAddr().String(), "error", err)
			// Remove from the connections
			removeIds = append(removeIds, id)
		}

	}
	lock.Unlock()

	for _, id := range removeIds {
		Remove(id)
	}
}

func SendTo(b []byte, ids ...ConnectionId) {
	lock.Lock()

	removeIds := []ConnectionId{}

	sentCt := 0
	// iterate through all provided id's and attempt to send

	for _, id := range ids {

		if cd, ok := netConnections[id]; ok {

			if _, err := cd.Write(b); err != nil {
				slog.Error("could not write to connection", "connectionId", id, "remoteAddr", cd.RemoteAddr().String(), "error", err)
				// Remove from the connections
				removeIds = append(removeIds, id)
				continue
			}

		}

		sentCt++
	}

	if sentCt < 1 {
		//slog.Info("message sent to nobody", "message", strings.Replace(string(b), "\033", "ESC", -1))
	}

	lock.Unlock()

	for _, id := range removeIds {
		Remove(id)
	}
}

// make this more efficient later
func ActiveConnectionCount() int {
	lock.RLock()
	defer lock.RUnlock()

	return len(netConnections)
}

// make this more efficient later
func SetShutdownChan(osSignalChan chan os.Signal) {
	lock.Lock()
	defer lock.Unlock()

	if shutdownChannel != nil {
		panic("Can't set shutdown channel a second time!")
	}
	shutdownChannel = osSignalChan
}

func Stats() (connections uint64, disconnections uint64) {
	lock.RLock()
	defer lock.RUnlock()

	return connectCounter, disconnectCounter
}

func GetClientSettings(id ConnectionId) ClientSettings {
	lock.Lock()
	defer lock.Unlock()

	if cd, ok := netConnections[id]; ok {
		return cd.clientSettings
	}

	return ClientSettings{}
}

func OverwriteClientSettings(id ConnectionId, cs ClientSettings) {
	lock.Lock()
	defer lock.Unlock()

	if cd, ok := netConnections[id]; ok {
		cd.clientSettings = cs
	}
}
