package scripting

import (
	"time"

	"github.com/dop251/goja"
)

type VMWrapper struct {
	VM            *goja.Runtime
	LastUsed      time.Time
	callableCache map[string]goja.Callable
	cacheSize     int
	maxCacheSize  int
}

func newVMWrapper(vm *goja.Runtime, cacheSize int) *VMWrapper {
	return &VMWrapper{VM: vm, LastUsed: time.Now(), callableCache: make(map[string]goja.Callable, cacheSize), maxCacheSize: cacheSize}
}

func (vmw *VMWrapper) GetFunction(name string) (goja.Callable, bool) {

	fn, ok := vmw.callableCache[name]

	if ok {
		return fn, fn != nil
	}

	fn, ok = goja.AssertFunction(vmw.VM.Get(name))

	if vmw.cacheSize < vmw.maxCacheSize {
		vmw.cacheSize++
		vmw.callableCache[name] = fn
	}

	return fn, ok
}

func (vmw *VMWrapper) MarkUsed() {
	vmw.LastUsed = time.Now()
}
