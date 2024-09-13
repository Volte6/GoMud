package scripting

import (
	"github.com/dop251/goja"
)

var (
	disableMessageQueue = false
)

func setMessagingFunctions(vm *goja.Runtime) {

	vm.Set(`console`, newConsole(vm))
	vm.Set(`SendUserMessage`, SendUserMessage)
	vm.Set(`SendRoomMessage`, SendRoomMessage)
	vm.Set(`SendRoomExitsMessage`, SendRoomExitsMessage)
	vm.Set(`SendBroadcast`, SendBroadcast)

}
