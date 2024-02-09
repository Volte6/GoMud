package scripting

import (
	"testing"

	"github.com/dop251/goja"
)

const (
	TEST_SCRIPT = `function TestFound() {}`
)

func Benchmark_Goja_Assert_Found(b *testing.B) {
	// Set up the VM
	vm := goja.New()
	vm.RunString(TEST_SCRIPT)

	for n := 0; n < b.N; n++ {
		goja.AssertFunction(vm.Get(`TestFound`))
	}
}

func Benchmark_VMW_Get_Found_Cached(b *testing.B) {
	// Set up the VM
	vm := goja.New()
	vm.RunString(TEST_SCRIPT)
	vmw := newVMWrapper(vm, 100)

	for n := 0; n < b.N; n++ {
		vmw.GetFunction(`TestFound`)
	}
}

func Benchmark_Goja_Assert_Missing(b *testing.B) {
	// Set up the VM
	vm := goja.New()
	vm.RunString(TEST_SCRIPT)

	for n := 0; n < b.N; n++ {
		goja.AssertFunction(vm.Get(`TestMissing`))
	}
}

func Benchmark_VMW_Get_Missing_Cached(b *testing.B) {
	// Set up the VM
	vm := goja.New()
	vm.RunString(TEST_SCRIPT)
	vmw := newVMWrapper(vm, 100)

	for n := 0; n < b.N; n++ {
		vmw.GetFunction(`TestMissing`)
	}
}

func Benchmark_VMW_Get_Found_NotCached(b *testing.B) {
	// Set up the VM
	vm := goja.New()
	vm.RunString(TEST_SCRIPT)
	vmw := newVMWrapper(vm, 0)

	for n := 0; n < b.N; n++ {
		vmw.GetFunction(`TestFound`)
	}
}

func Benchmark_VMW_Get_Missing_NotCached(b *testing.B) {
	// Set up the VM
	vm := goja.New()
	vm.RunString(TEST_SCRIPT)
	vmw := newVMWrapper(vm, 0)

	for n := 0; n < b.N; n++ {
		vmw.GetFunction(`TestMissing`)
	}
}
