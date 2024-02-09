package scripting

import (
	"log/slog"

	"github.com/dop251/goja"
)

type console struct{}

func (c *console) log(msg any) {
	slog.Info("JSVM", "msg", msg)
}
func (c *console) info(msg any) {
	slog.Info("JSVM", "msg", msg)
}
func (c *console) debug(msg any) {
	slog.Debug("JSVM", "msg", msg)
}
func (c *console) warn(msg any) {
	slog.Warn("JSVM", "msg", msg)
}
func (c *console) error(msg any) {
	slog.Error("JSVM", "msg", msg)
}

func newConsole(vm *goja.Runtime) *goja.Object {
	c := &console{}
	obj := vm.NewObject()
	obj.Set("log", c.log)
	obj.Set("info", c.info)
	obj.Set("debug", c.debug)
	obj.Set("warn", c.warn)
	obj.Set("error", c.error)
	return obj
}
