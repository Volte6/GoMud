package connection

import (
	"errors"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type ConnectState uint32

const (
	Login ConnectState = iota
	Zombie
	MaxHistory = 10
)

type ClientSettings struct {
	ScreenWidth  uint32
	ScreenHeight uint32
	Monochrome   bool
}

type InputHistory struct {
	inhistory bool
	position  int
	history   [][]byte
}

func (ih *InputHistory) Get() []byte {
	if len(ih.history) < 1 {
		return nil
	}

	return ih.history[ih.position]
}

func (ih *InputHistory) Add(input []byte) {

	if len(ih.history) >= MaxHistory {
		ih.history = ih.history[1:]
	}

	ih.history = append(ih.history, make([]byte, len(input)))
	ih.position = len(ih.history) - 1
	copy(ih.history[ih.position], input)
	ih.inhistory = false
}

func (ih *InputHistory) Previous() {
	if !ih.inhistory {
		ih.inhistory = true
		return
	}
	if ih.position <= 0 {
		return
	}

	ih.position--
}

func (ih *InputHistory) Next() {
	if !ih.inhistory {
		ih.inhistory = true
		return
	}
	if ih.position >= len(ih.history)-1 {
		return
	}

	ih.position++
}

// returns position and whether position is not the last item
func (ih *InputHistory) Position() int {
	return ih.position
}

func (ih *InputHistory) ResetPosition() {
	ih.inhistory = false
	ih.position = len(ih.history) - 1
	if ih.position < 0 {
		ih.position = 0
	}
}

func (ih *InputHistory) InHistory() bool {
	return ih.inhistory
}

// A structure to package up everything we need to know about this input.
type ClientInput struct {
	ConnectionId  ConnectionId // Who does this belong to?
	DataIn        []byte       // What was the last thing they typed?
	Buffer        []byte       // What is the current buffer
	Clipboard     []byte       // Text that can be easily pasted with ctrl-v
	LastSubmitted []byte       // The last thing submitted
	EnterPressed  bool         // Did they hit enter? It's stripped from the buffer/input FYI
	BSPressed     bool         // Did they hit backspace?
	TabPressed    bool         // Did they hit tab?
	History       InputHistory // A list of the last 10 things they typed
}

// Reset the client input to essentially "No current input"
func (ci *ClientInput) Reset() {
	ci.DataIn = ci.DataIn[:0]
	ci.Buffer = ci.Buffer[:0]
	ci.EnterPressed = false
}

type InputHandler func(ci *ClientInput, ct *ConnectionTracker, handlerState map[string]any) (doNextHandler bool)

type ConnectionDetails struct {
	connectionId      ConnectionId
	state             ConnectState
	lastInputTime     time.Time
	settings          ClientSettings
	settingsMutex     sync.Mutex
	conn              net.Conn
	handlerMutex      sync.Mutex
	inputHandlerNames []string
	inputHandlers     []InputHandler
}

// bool is true if they have changed since last time they were gotten
func (cd *ConnectionDetails) GetSettings() ClientSettings {
	cd.settingsMutex.Lock()
	defer cd.settingsMutex.Unlock()
	return cd.settings
}

// If HandleInput receives an error, we shouldn't pass input to the game logic
func (cd *ConnectionDetails) HandleInput(ci *ClientInput, ct *ConnectionTracker, handlerState map[string]any) (doNextHandler bool, lastHandler string, err error) {
	cd.handlerMutex.Lock()
	defer cd.handlerMutex.Unlock()

	cd.lastInputTime = time.Now()

	handlerCt := len(cd.inputHandlers)
	if handlerCt < 1 {
		return false, lastHandler, errors.New("no input handlers")
	}

	for i, inputHandler := range cd.inputHandlers {
		lastHandler = cd.inputHandlerNames[i]
		if runNextHandler := inputHandler(ci, ct, handlerState); !runNextHandler {
			// If it's the last one in the chain, ignore any aborts
			// if i == handlerCt-1 {
			// 	return false, lastHandler, nil
			// }
			return false, lastHandler, nil
		}
	}
	return true, lastHandler, nil
}

func (cd *ConnectionDetails) RemoveInputHandler(name string) {
	cd.handlerMutex.Lock()
	defer cd.handlerMutex.Unlock()

	for i := len(cd.inputHandlerNames) - 1; i >= 0; i-- {
		if cd.inputHandlerNames[i] == name {
			cd.inputHandlerNames = append(cd.inputHandlerNames[:i], cd.inputHandlerNames[i+1:]...)
			cd.inputHandlers = append(cd.inputHandlers[:i], cd.inputHandlers[i+1:]...)
		}
	}

}

func (cd *ConnectionDetails) AddInputHandler(name string, newInputHandler InputHandler, after ...string) {
	cd.handlerMutex.Lock()
	defer cd.handlerMutex.Unlock()

	if len(after) > 0 {
		for i, handlerName := range cd.inputHandlerNames {
			if handlerName == after[0] {
				cd.inputHandlerNames = append(cd.inputHandlerNames[:i+1], append([]string{name}, cd.inputHandlerNames[i+1:]...)...)
				cd.inputHandlers = append(cd.inputHandlers[:i+1], append([]InputHandler{newInputHandler}, cd.inputHandlers[i+1:]...)...)
				return
			}
		}
	}

	cd.inputHandlerNames = append(cd.inputHandlerNames, name)
	cd.inputHandlers = append(cd.inputHandlers, newInputHandler)
}

func (cd *ConnectionDetails) Write(p []byte) (n int, err error) {

	p = []byte(strings.ReplaceAll(string(p), "\n", "\r\n"))

	return cd.conn.Write(p)
}

func (cd *ConnectionDetails) Read(p []byte) (n int, err error) {
	return cd.conn.Read(p)
}

func (cd *ConnectionDetails) Close() {
	cd.conn.Close()
}

func (cd *ConnectionDetails) RemoteAddr() net.Addr {
	return cd.conn.RemoteAddr()
}

// get for uniqueId
func (cd *ConnectionDetails) ConnectionId() ConnectionId {
	return ConnectionId(atomic.LoadUint64((*uint64)(&cd.connectionId)))
}

// set and get for state
func (cd *ConnectionDetails) State() ConnectState {
	return ConnectState(atomic.LoadUint32((*uint32)(&cd.state)))
}
func (cd *ConnectionDetails) SetState(state ConnectState) {
	atomic.StoreUint32((*uint32)(&cd.state), uint32(state))
}

func (cd *ConnectionDetails) SetScreenSize(w uint32, h uint32) {

	cd.settingsMutex.Lock()
	defer cd.settingsMutex.Unlock()

	cd.settings.ScreenWidth = w
	cd.settings.ScreenHeight = h
}

func (cd *ConnectionDetails) SetMonochrome(mono bool) {
	cd.settingsMutex.Lock()
	defer cd.settingsMutex.Unlock()

	cd.settings.Monochrome = mono
}

func NewConnectionDetails(connId ConnectionId, c net.Conn) *ConnectionDetails {
	return &ConnectionDetails{
		state:        Login,
		connectionId: connId,
		settings: ClientSettings{
			ScreenWidth:  80,
			ScreenHeight: 24,
			Monochrome:   false,
		},
		conn: c,
	}
}
