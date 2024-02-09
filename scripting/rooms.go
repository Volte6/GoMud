package scripting

import (
	"github.com/dop251/goja"
)

type rooms struct{}

func (r *rooms) LoadRoom(roomId int) *Room {
	return rooms.LoadRoom(user.Character.RoomId)
}

func NewRoom(vm *goja.Runtime) *goja.Object {
	c := &rooms{}
	obj := vm.NewObject()
	obj.Set("log", c.log)
	obj.Set("info", c.info)
	obj.Set("debug", c.debug)
	obj.Set("warn", c.warn)
	obj.Set("error", c.error)
	return obj
}
