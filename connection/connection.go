package connection

import (
	"errors"
	"log/slog"
	"net"
	"os"
	"sync"
	"sync/atomic"
)

const ReadBufferSize = 1024

type ConnectionId = uint64

type ConnectionTracker struct {
	connectCounter    uint64         // a counter for each time a connection is accepted
	disconnectCounter uint64         // a counter for each tim ea connection is dropped
	netConnections    sync.Map       // a mapping of unique id's to connections
	shutdownChannel   chan os.Signal // channel to receive shutdown signals
}

func (c *ConnectionTracker) Signal(s os.Signal) {
	c.shutdownChannel <- s
}

func (c *ConnectionTracker) Add(conn net.Conn) *ConnectionDetails {

	uId := ConnectionId(atomic.AddUint64(&c.connectCounter, 1))

	connDetails := NewConnectionDetails(
		uId,
		conn,
	)

	c.netConnections.Store(connDetails.ConnectionId(), connDetails)

	// return the unique ID to find this connection later
	return connDetails
}

// Returns the total number of connections
func (c *ConnectionTracker) Get(id ConnectionId) (cd *ConnectionDetails, err error) {

	// Try to retrieve the value
	if cd, ok := c.netConnections.Load(id); ok {
		return cd.(*ConnectionDetails), nil
	}

	return nil, errors.New("connection not found")
}

func (c *ConnectionTracker) Cleanup() {
	c.netConnections.Range(func(key, value interface{}) (processNext bool) {
		c.Remove(key.(ConnectionId))
		return true
	})
}

func (c *ConnectionTracker) Remove(id ConnectionId) (err error) {

	//if err := users.LogOutUserByConnectionId(id); err != nil {
	//	slog.Error("could not log out user", "connectionId", id, "error", err)
	//}

	// Try to retrieve the value
	if cd, ok := c.netConnections.Load(id); ok {
		// close the connection, no longer useful.
		cd.(*ConnectionDetails).Close()
		// keep track of the number of disconnects
		atomic.AddUint64(&c.disconnectCounter, 1)
		// remove the connection from the map
		c.netConnections.Delete(id)

		slog.Info("connection removed", "connectionId", id, "remoteAddr", cd.(*ConnectionDetails).RemoteAddr().String())
		return nil
	}

	return errors.New("connection not found")
}

func (c *ConnectionTracker) Broadcast(b []byte, excludeIds ...ConnectionId) {

	excludeLen := len(excludeIds)

	// Range over the sync.Map
	c.netConnections.Range(func(key, cd interface{}) (processNext bool) {

		// If in the exclusion list, skip it
		if excludeLen > 0 && sliceContains(excludeIds, key.(ConnectionId)) {
			return true
		}

		details := cd.(*ConnectionDetails)
		if details.state != Login {
			// Write the message to the connection
			if _, err := details.Write(b); err != nil {
				slog.Error("could not write to connection", "connectionId", key.(uint64), "remoteAddr", cd.(*ConnectionDetails).RemoteAddr().String(), "error", err)
				// Remove from the connections
				c.Remove(key.(ConnectionId))
			}
		}

		return true // return true unless you want it to halt early
	})

}

func (c *ConnectionTracker) SendTo(b []byte, ids ...ConnectionId) {

	sentCt := 0
	// iterate through all provided id's and attempt to send
	for _, id := range ids {
		if cd, ok := c.netConnections.Load(id); ok {
			if _, err := cd.(*ConnectionDetails).Write(b); err != nil {
				slog.Error("could not write to connection", "connectionId", id, "remoteAddr", cd.(*ConnectionDetails).RemoteAddr().String(), "error", err)
				// Remove from the connections
				c.Remove(id)
			}
			sentCt++
		}
	}

	if sentCt < 1 {
		slog.Info("message sent to nobody", "message", string(b))
	}
}

func sliceContains(slice []ConnectionId, id ConnectionId) bool {
	for _, v := range slice {
		if v == id {
			return true
		}
	}
	return false
}

// make this more efficient later
func (c *ConnectionTracker) ActiveConnectionCount() int {
	ct := 0
	c.netConnections.Range(func(key, cd interface{}) (processNext bool) {
		ct++
		return true
	})
	return ct

}

func (c *ConnectionTracker) Stats() (connections uint64, disconnections uint64) {
	return atomic.LoadUint64(&c.connectCounter), atomic.LoadUint64(&c.disconnectCounter)
}

var connTracker *ConnectionTracker = nil

func New(osSignalChan chan os.Signal) *ConnectionTracker {
	return &ConnectionTracker{shutdownChannel: osSignalChan}
}
