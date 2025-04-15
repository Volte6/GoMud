package scripting

import (
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/dop251/goja"
)

type console struct{}

func (c *console) log(msg any) {
	mudlog.Info(`JSVM`, `msg`, msg)
}
func (c *console) info(msg any) {
	mudlog.Info(`JSVM`, `msg`, msg)
}
func (c *console) debug(msg any) {
	mudlog.Debug(`JSVM`, `msg`, msg)
}
func (c *console) warn(msg any) {
	mudlog.Warn(`JSVM`, `msg`, msg)
}
func (c *console) error(msg any) {
	mudlog.Error(`JSVM`, `msg`, msg)
}

func newConsole(vm *goja.Runtime) *goja.Object {
	c := &console{}
	obj := vm.NewObject()
	obj.Set(`log`, c.log)
	obj.Set(`info`, c.info)
	obj.Set(`debug`, c.debug)
	obj.Set(`warn`, c.warn)
	obj.Set(`error`, c.error)
	return obj
}
