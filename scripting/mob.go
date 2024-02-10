package scripting

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/dop251/goja"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/util"
)

var (
	mobVMCache       = make(map[int]*VMWrapper)
	scriptMobTimeout = 10 * time.Millisecond
)

func PruneMobVMs(instanceIds ...int) {
	if len(instanceIds) > 0 {
		for _, mobInstanceId := range instanceIds {
			delete(mobVMCache, mobInstanceId)
		}
		return
	}
	for mobInstanceId, _ := range mobVMCache {
		if mob := mobs.GetInstance(mobInstanceId); mob == nil {
			delete(mobVMCache, mobInstanceId)
		}
	}
}

func TryMobScriptEvent(eventName string, mobInstanceId int, roomId int, sourceId int, sourceType string, details map[string]any, cmdQueue util.CommandQueue) (util.MessageQueue, error) {
	timestart := time.Now()
	defer func() {
		slog.Debug("TryMobScriptEvent()", "time", time.Since(timestart))
	}()

	messageQueue = util.NewMessageQueue(0, mobInstanceId)
	commandQueue = cmdQueue

	vmw, err := getMobVM(mobInstanceId)
	if err != nil {
		return messageQueue, err
	}
	slog.Info("TryMobScriptEvent()", "eventName", eventName, "mobInstanceId", mobInstanceId, "roomId", roomId, "details", details)
	if onCommandFunc, ok := vmw.GetFunction(eventName); ok {

		tmr := time.AfterFunc(scriptRoomTimeout, func() {
			vmw.VM.Interrupt(errTimeout)
		})

		if details == nil {
			details = make(map[string]any)
		}

		details["sourceId"] = sourceId
		details["sourceType"] = sourceType

		res, err := onCommandFunc(goja.Undefined(),
			vmw.VM.ToValue(mobInstanceId),
			vmw.VM.ToValue(roomId),
			vmw.VM.ToValue(details),
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

func TryMobCommand(cmd string, rest string, mobInstanceId int, sourceId int, sourceType string, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	timestart := time.Now()
	defer func() {
		slog.Debug("TryMobCommand()", "time", time.Since(timestart))
	}()

	messageQueue = util.NewMessageQueue(0, mobInstanceId)
	commandQueue = cmdQueue

	mob := mobs.GetInstance(mobInstanceId)
	if mob == nil {
		PruneMobVMs(mobInstanceId)
		return messageQueue, errors.New("mob not found")
	}

	vmw, err := getMobVM(mobInstanceId)
	if err != nil {
		return messageQueue, err
	}

	if onCommandFunc, ok := vmw.GetFunction(`onCommand_` + cmd); ok {

		details := map[string]interface{}{
			`sourceId`:   sourceId,
			`sourceType`: sourceType,
		}

		tmr := time.AfterFunc(scriptRoomTimeout, func() {
			vmw.VM.Interrupt(errTimeout)
		})
		res, err := onCommandFunc(goja.Undefined(),
			vmw.VM.ToValue(rest),
			vmw.VM.ToValue(mobInstanceId),
			vmw.VM.ToValue(mob.Character.RoomId),
			vmw.VM.ToValue(details),
		)
		vmw.VM.ClearInterrupt()
		tmr.Stop()

		if err != nil {

			// Wrap the error
			finalErr := fmt.Errorf("onCommand_%s(): %w", cmd, err)

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

	} else if onCommandFunc, ok := vmw.GetFunction(`onCommand`); ok {

		details := map[string]interface{}{
			`sourceId`:   sourceId,
			`sourceType`: sourceType,
		}

		tmr := time.AfterFunc(scriptRoomTimeout, func() {
			vmw.VM.Interrupt(errTimeout)
		})
		res, err := onCommandFunc(goja.Undefined(),
			vmw.VM.ToValue(cmd),
			vmw.VM.ToValue(rest),
			vmw.VM.ToValue(mobInstanceId),
			vmw.VM.ToValue(mob.Character.RoomId),
			vmw.VM.ToValue(details),
		)
		vmw.VM.ClearInterrupt()
		tmr.Stop()

		if err != nil {

			// Wrap the error
			finalErr := fmt.Errorf("onCommand(): %w", err)

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

func getMobVM(mobInstanceId int) (*VMWrapper, error) {

	if vm, ok := mobVMCache[mobInstanceId]; ok {
		mobVMCache[mobInstanceId] = vm
		if vm == nil {
			return nil, errNoScript
		}
		return vm, nil
	}

	mob := mobs.GetInstance(mobInstanceId)
	if mob == nil {
		return nil, fmt.Errorf("mob instance not found: %d", mobInstanceId)
	}

	script := mob.GetScript()
	if len(script) == 0 {
		mobVMCache[mobInstanceId] = nil
		return nil, errNoScript
	}

	vm := goja.New()
	setAllScriptingFunctions(vm)

	prg, err := goja.Compile(fmt.Sprintf(`mob-%d`, mobInstanceId), script, false)
	if err != nil {
		finalErr := fmt.Errorf("Compile: %w", err)
		return nil, finalErr
	}

	//
	// Run the program
	//
	tmr := time.AfterFunc(scriptMobTimeout, func() {
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

	//
	// Run onLoad() function
	//
	tmr = time.AfterFunc(scriptLoadTimeout, func() {
		vm.Interrupt(errTimeout)
	})
	if fn, ok := goja.AssertFunction(vm.Get(`onLoad`)); ok {
		if _, err := fn(goja.Undefined(), vm.ToValue(mobInstanceId)); err != nil {
			// Wrap the error
			finalErr := fmt.Errorf("onLoad: %w", err)

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
	}
	vm.ClearInterrupt()
	tmr.Stop()

	vmw := newVMWrapper(vm, 100)

	mobVMCache[mobInstanceId] = vmw

	return vmw, nil
}
