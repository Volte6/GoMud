package scripting

import (
	"errors"
	"time"

	"github.com/dop251/goja"
)

var (
	errNoScript = errors.New("no script")
	errTimeout  = errors.New("script timeout")
)

func Setup(scriptLoadTimeoutMs int, scriptRoomTimeoutMs int) {
	scriptLoadTimeout = time.Duration(scriptLoadTimeoutMs) * time.Millisecond
	scriptRoomTimeout = time.Duration(scriptRoomTimeoutMs) * time.Millisecond
}

func setAllScriptingFunctions(vm *goja.Runtime) {
	setMessagingFunctions(vm)
	setRoomFunctions(vm)
	setActorFunctions(vm)
	setUtilFunctions(vm)
}

func PruneVMs() {
	PruneRoomVMs()
	PruneMobVMs()
}
