package scripting

import (
	"github.com/dop251/goja"
	"github.com/volte6/mud/util"
)

var (
	disableMessageQueue = false
	messageQueue        util.MessageQueue
	commandQueue        util.CommandQueue
)

func setMessagingFunctions(vm *goja.Runtime) {

	vm.Set(`console`, newConsole(vm))
	vm.Set(`SendUserMessage`, SendUserMessage)
	vm.Set(`SendRoomMessage`, SendRoomMessage)
	vm.Set(`SendBroadcast`, SendBroadcast)

}
