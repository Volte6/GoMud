package scripting

import (
	"fmt"
	"log/slog"
	"reflect"

	"github.com/dop251/goja"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

const (
	TEST_SCRIPT = `
	function onCommand(cmd, rest, user, response) {

		console.log("1onCommand-"+cmd+"/rest-"+rest+"/userId-"+String(user.UserId))
		

		if (cmd === "ping") {
			response.SendUserMessage(userId, "Oops!", true)
			response.Handled = true;
		}
		// throw("Test");
		return response;
	}`

	TEST_SCRIPT2 = `
	function onCommand(cmd, rest, user, response) {

		console.log("2onCommand-"+cmd+"/rest-"+rest+"/userId-"+String(user.UserId))
		user.Character.Namio = "Chuckles"
		
		if (cmd === "ping") {
			response.SendUserMessage(userId, "Oops!", true)
			response.Handled = true;
		}
		// throw("Test");
		return response;
	}`
)

func Test(u *users.UserRecord) {

	vm := goja.New()
	vm.Set("console", newConsole(vm))

	_, err := vm.RunString(TEST_SCRIPT)
	if err != nil {
		panic(err)
	}

	_, err = vm.RunString(TEST_SCRIPT2)
	if err != nil {
		panic(err)
	}

	onCommandFunc, ok := goja.AssertFunction(vm.Get("onCommand"))
	if !ok {
		panic("Not a function")
	}

	cmd := "ping"
	rest := "north"

	res, err := onCommandFunc(goja.Undefined(),
		vm.ToValue(cmd),
		vm.ToValue(rest),
		vm.ToValue(u),
		vm.ToValue(util.NewMessageQueue(u.UserId, 0)),
		vm.ToValue("test"))

	if err != nil {

		if jserr, ok := err.(*goja.Exception); ok {
			slog.Error("JSVM", "exception", jserr.Value().Export())
		} else {
			panic(err)
		}

	} else {

		mQ := res.Export().(util.MessageQueue)
		fmt.Println("Handled", mQ.Handled)

	}

	slog.Info("NAme", u.Character.Name)
}

func sizeOf(v reflect.Value) uintptr {
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return 0
		}
		return sizeOf(v.Elem())

	case reflect.Slice:
		if v.IsNil() {
			return 0
		}
		length := v.Len()
		elemSize := sizeOf(reflect.New(v.Type().Elem()).Elem())
		return uintptr(length) * elemSize

	case reflect.Struct:
		var size uintptr
		for i := 0; i < v.NumField(); i++ {
			size += sizeOf(v.Field(i))
		}
		return size

	case reflect.Array:
		length := v.Len()
		elemSize := sizeOf(reflect.New(v.Type().Elem()).Elem())
		return uintptr(length) * elemSize

	case reflect.String:
		return uintptr(len(v.String()))

	case reflect.Map:
		// Maps are tricky because they have an unknown overhead for buckets and other internals.
		// A rough estimate is the size of the keys and values, but this omits the actual map overhead.
		// You might add a constant factor or use a per-map overhead based on runtime/map.go info.
		var size uintptr
		keys := v.MapKeys()
		for _, key := range keys {
			size += sizeOf(key) + sizeOf(v.MapIndex(key))
		}
		return size

	default:
		// This accounts for the types like integers, bools, etc.
		return v.Type().Size()
	}
}
