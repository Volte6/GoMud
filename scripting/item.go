package scripting

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/dop251/goja"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/util"
)

var (
	itemVMCache       = make(map[int]*VMWrapper)
	scriptItemTimeout = 10 * time.Millisecond
)

func PruneItemVMs(instanceIds ...int) {
	// Do not prune, they dont' get a VM per buff instance.
}

func TryItemScriptEvent(eventName string, userId int, mobInstanceId int, item items.Item, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	vmw, err := getItemVM(item.ItemId)
	if err != nil {
		return util.NewMessageQueue(0, 0), err
	}

	messageQueue = util.NewMessageQueue(0, 0)
	commandQueue = cmdQueue

	actorInfo := GetActor(userId, mobInstanceId)

	timestart := time.Now()
	defer func() {
		slog.Debug("TryItemScriptEvent()", "eventName", eventName, "itemId", item.ItemId, "time", time.Since(timestart))
	}()
	if onCommandFunc, ok := vmw.GetFunction(eventName); ok {

		tmr := time.AfterFunc(scriptRoomTimeout, func() {
			vmw.VM.Interrupt(errTimeout)
		})

		res, err := onCommandFunc(goja.Undefined(),
			vmw.VM.ToValue(item),
			vmw.VM.ToValue(actorInfo),
		)
		vmw.VM.ClearInterrupt()
		tmr.Stop()

		if err != nil {

			// Wrap the error
			finalErr := fmt.Errorf("%s(): %w", eventName, err)

			if _, ok := finalErr.(*goja.Exception); ok {
				slog.Error("JSVM", "exception", finalErr)
				return messageQueue, finalErr
			} else if errors.Is(finalErr, errTimeout) {
				slog.Error("JSVM", "interrupted", finalErr)
				return messageQueue, finalErr
			}

			slog.Error("JSVM", "error", finalErr)
			return messageQueue, finalErr
		}

		if boolVal, ok := res.Export().(bool); ok {
			messageQueue.Handled = messageQueue.Handled || boolVal
		}
	}

	return messageQueue, nil
}

func getItemVM(itemId int) (*VMWrapper, error) {

	if vm, ok := itemVMCache[itemId]; ok {
		if vm == nil {
			return nil, errNoScript
		}
		return vm, nil
	}

	spec := items.GetItemSpec(itemId)

	script := spec.GetScript()
	if len(script) == 0 {
		itemVMCache[itemId] = nil
		return nil, errNoScript
	}

	vm := goja.New()
	setAllScriptingFunctions(vm)

	prg, err := goja.Compile(fmt.Sprintf(`item-%d`, itemId), script, false)
	if err != nil {
		finalErr := fmt.Errorf("Compile: %w", err)
		return nil, finalErr
	}

	//
	// Run the program
	//
	tmr := time.AfterFunc(scriptLoadTimeout, func() {
		vm.Interrupt(errTimeout)
	})
	if _, err = vm.RunProgram(prg); err != nil {

		// Wrap the error
		finalErr := fmt.Errorf("RunProgram: %w", err)

		if _, ok := finalErr.(*goja.Exception); ok {
			slog.Error("JSVM", "exception", finalErr)
			return nil, finalErr
		} else if errors.Is(finalErr, errTimeout) {
			slog.Error("JSVM", "interrupted", finalErr)
			return nil, finalErr
		}

		slog.Error("JSVM", "error", finalErr)
		return nil, finalErr
	}
	vm.ClearInterrupt()
	tmr.Stop()

	vmw := newVMWrapper(vm, 0)

	itemVMCache[itemId] = vmw

	return vmw, nil
}
