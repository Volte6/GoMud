package connections

import (
	"errors"
	"github.com/gorilla/websocket"
	"log/slog"
	"sync"
	"time"
)

type HeartbeatConfig struct {
	PongWait   time.Duration
	PingPeriod time.Duration
	WriteWait  time.Duration
}

var DefaultHeartbeatConfig = HeartbeatConfig{
	PongWait:   60 * time.Second,
	PingPeriod: (60 * time.Second * 9) / 10, // Must be less than PongWait, 90% seems to be common
	WriteWait:  10 * time.Second,
}

var (
	ErrNotWebsocket = errors.New("connection is not a websocket")
	ErrWriteFailed  = errors.New("failed to write message")
)

type heartbeatManager struct {
	cd       *ConnectionDetails
	config   HeartbeatConfig
	stopChan chan struct{}
	wg       sync.WaitGroup
}

func newHeartbeatManager(cd *ConnectionDetails, config HeartbeatConfig) *heartbeatManager {
	return &heartbeatManager{
		cd:       cd,
		config:   config,
		stopChan: make(chan struct{}),
	}
}

func (cd *ConnectionDetails) StartHeartbeat(config HeartbeatConfig) error {
	if cd.wsConn == nil {
		return ErrNotWebsocket
	}

	hm := newHeartbeatManager(cd, config)
	slog.Info("Heartbeat::Start", "connectionId", cd.connectionId)
	// set up pong handler
	cd.wsConn.SetReadDeadline(time.Now().Add(hm.config.PongWait))
	cd.wsConn.SetPongHandler(func(string) error {
		slog.Debug("Heartbeat::Pong", "connectionId", hm.cd.connectionId)
		cd.wsConn.SetReadDeadline(time.Now().Add(hm.config.PongWait))
		return nil
	})

	// start ping ticker in a goroutine
	hm.wg.Add(1)
	go hm.runPingLoop()

	return nil
}

func (hm *heartbeatManager) runPingLoop() {
	defer hm.wg.Done()
	ticker := time.NewTicker(hm.config.PingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-hm.stopChan:
			return
		case <-ticker.C:
			if err := hm.writePing(); err != nil {
				slog.Error("Heartbeat::Error",
					"connectionId", hm.cd.connectionId,
					"error", err)
				return
			}
		}
	}
}

func (hm *heartbeatManager) writePing() error {
	hm.cd.wsLock.Lock()
	defer hm.cd.wsLock.Unlock()

	deadline := time.Now().Add(hm.config.WriteWait)
	slog.Debug("Heartbeat::Ping", "connectionId", hm.cd.connectionId)

	if err := hm.cd.wsConn.WriteControl(
		websocket.PingMessage,
		nil,
		deadline); err != nil {
		if !errors.Is(err, websocket.ErrCloseSent) {
			return errors.Join(ErrWriteFailed, err)
		}
		return err
	}
	return nil
}

func (hm *heartbeatManager) stop() {
	close(hm.stopChan)
	hm.wg.Wait()
}
