package scripting

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/dop251/goja"
)

var (
	mobVMCache       = make(map[string]*VMWrapper)
	scriptMobTimeout = 50 * time.Millisecond
)

func ClearMobVMs() {
	clear(mobVMCache)
}

func PruneMobVMs(instanceIds ...int) {

}

func TryMobConverse(rest string, mobInstanceId int, sourceMobInstanceId int) (bool, error) {

	sMob := GetActor(0, mobInstanceId)
	if sMob == nil {
		return false, errors.New("mob not found")
	}

	vmw, err := getMobVM(sMob)
	if err != nil {
		return false, err
	}

	timestart := time.Now()
	defer func() {
		slog.Debug("TryMobConverse()", "mobInstanceId", mobInstanceId, "sourceMobInstanceId", sourceMobInstanceId, "time", time.Since(timestart))
	}()
	if onCommandFunc, ok := vmw.GetFunction("onConverse"); ok {

		tmr := time.AfterFunc(scriptRoomTimeout, func() {
			vmw.VM.Interrupt(errTimeout)
		})

		sourceMob := GetActor(0, sourceMobInstanceId)
		if sourceMob == nil {
			return false, errors.New("mob not found")
		}

		sRoom := GetRoom(sMob.GetRoomId())

		res, err := onCommandFunc(goja.Undefined(),
			vmw.VM.ToValue(rest),
			vmw.VM.ToValue(sMob),
			vmw.VM.ToValue(sourceMob),
			vmw.VM.ToValue(sRoom),
		)
		vmw.VM.ClearInterrupt()
		tmr.Stop()

		if err != nil {

			// Wrap the error
			finalErr := fmt.Errorf("TryMobConverse(): %w", err)

			if _, ok := finalErr.(*goja.Exception); ok {
				slog.Error("JSVM", "exception", finalErr)
				return false, finalErr
			} else if errors.Is(finalErr, errTimeout) {
				slog.Error("JSVM", "interrupted", finalErr)
				return false, finalErr
			}

			slog.Error("JSVM", "error", finalErr)
			return false, finalErr
		}

		if boolVal, ok := res.Export().(bool); ok {
			return boolVal, nil
		}
	}

	return false, nil
}

func TryMobScriptEvent(eventName string, mobInstanceId int, sourceId int, sourceType string, details map[string]any) (bool, error) {

	sMob := GetActor(0, mobInstanceId)
	if sMob == nil {
		return false, errors.New("mob not found")
	}

	vmw, err := getMobVM(sMob)
	if err != nil {
		return false, err
	}

	timestart := time.Now()
	defer func() {
		slog.Debug("TryMobScriptEvent()", "eventName", eventName, "MobId", sMob.MobTypeId(), "time", time.Since(timestart))
	}()
	if onCommandFunc, ok := vmw.GetFunction(eventName); ok {

		tmr := time.AfterFunc(scriptRoomTimeout, func() {
			vmw.VM.Interrupt(errTimeout)
		})

		if details == nil {
			details = make(map[string]any)
		}

		sRoom := GetRoom(sMob.GetRoomId())

		details["sourceId"] = sourceId
		details["sourceType"] = sourceType

		res, err := onCommandFunc(goja.Undefined(),
			vmw.VM.ToValue(sMob),
			vmw.VM.ToValue(sRoom),
			vmw.VM.ToValue(details),
		)
		vmw.VM.ClearInterrupt()
		tmr.Stop()

		if err != nil {

			// Wrap the error
			finalErr := fmt.Errorf("%s(): %w", eventName, err)

			if _, ok := finalErr.(*goja.Exception); ok {
				slog.Error("JSVM", "exception", finalErr)
				return false, finalErr
			} else if errors.Is(finalErr, errTimeout) {
				slog.Error("JSVM", "interrupted", finalErr)
				return false, finalErr
			}

			slog.Error("JSVM", "error", finalErr)
			return false, finalErr
		}

		if boolVal, ok := res.Export().(bool); ok {
			return boolVal, nil
		}
	}

	return false, nil
}

func TryMobCommand(cmd string, rest string, mobInstanceId int, sourceId int, sourceType string) (bool, error) {

	sMob := GetActor(0, mobInstanceId)
	if sMob == nil {
		PruneMobVMs(mobInstanceId)
		return false, errors.New("mob not found")
	}

	vmw, err := getMobVM(sMob)
	if err != nil {
		return false, err
	}

	timestart := time.Now()
	defer func() {
		slog.Debug("TryMobCommand()", "cmd", cmd, "MobId", sMob.MobTypeId(), "time", time.Since(timestart))
	}()

	if onCommandFunc, ok := vmw.GetFunction(`onCommand_` + cmd); ok {

		details := map[string]interface{}{
			`sourceId`:   sourceId,
			`sourceType`: sourceType,
		}

		sRoom := GetRoom(sMob.mobRecord.Character.RoomId)

		tmr := time.AfterFunc(scriptRoomTimeout, func() {
			vmw.VM.Interrupt(errTimeout)
		})
		res, err := onCommandFunc(goja.Undefined(),
			vmw.VM.ToValue(rest),
			vmw.VM.ToValue(sMob),
			vmw.VM.ToValue(sRoom),
			vmw.VM.ToValue(details),
		)
		vmw.VM.ClearInterrupt()
		tmr.Stop()

		if err != nil {

			// Wrap the error
			finalErr := fmt.Errorf("onCommand_%s(): %w", cmd, err)

			if _, ok := finalErr.(*goja.Exception); ok {
				slog.Error("JSVM", "exception", finalErr)
				return false, finalErr
			} else if errors.Is(finalErr, errTimeout) {
				slog.Error("JSVM", "interrupted", finalErr)
				return false, finalErr
			}

			slog.Error("JSVM", "error", finalErr)
			return false, finalErr
		}

		if boolVal, ok := res.Export().(bool); ok {
			return boolVal, nil
		}

	} else if onCommandFunc, ok := vmw.GetFunction(`onCommand`); ok {

		details := map[string]interface{}{
			`sourceId`:   sourceId,
			`sourceType`: sourceType,
		}

		sRoom := GetRoom(sMob.GetRoomId())

		tmr := time.AfterFunc(scriptRoomTimeout, func() {
			vmw.VM.Interrupt(errTimeout)
		})
		res, err := onCommandFunc(goja.Undefined(),
			vmw.VM.ToValue(cmd),
			vmw.VM.ToValue(rest),
			vmw.VM.ToValue(sMob),
			vmw.VM.ToValue(sRoom),
			vmw.VM.ToValue(details),
		)
		vmw.VM.ClearInterrupt()
		tmr.Stop()

		if err != nil {

			// Wrap the error
			finalErr := fmt.Errorf("onCommand(): %w", err)

			if _, ok := finalErr.(*goja.Exception); ok {
				slog.Error("JSVM", "exception", finalErr)
				return false, finalErr
			} else if errors.Is(finalErr, errTimeout) {
				slog.Error("JSVM", "interrupted", finalErr)
				return false, finalErr
			}

			slog.Error("JSVM", "error", finalErr)
			return false, finalErr
		}

		if boolVal, ok := res.Export().(bool); ok {
			return boolVal, nil
		}
	}

	return false, nil
}

func getMobVM(mobActor *ScriptActor) (*VMWrapper, error) {

	scriptId := fmt.Sprintf(`%d-%s`, mobActor.MobTypeId(), mobActor.getScriptTag())

	if vm, ok := mobVMCache[scriptId]; ok {
		if vm == nil {
			return nil, errNoScript
		}
		return vm, nil
	}

	script := mobActor.getScript()
	if len(script) == 0 {
		mobVMCache[scriptId] = nil
		return nil, errNoScript
	}

	vm := goja.New()
	setAllScriptingFunctions(vm)

	prg, err := goja.Compile(fmt.Sprintf(`mob-%s`, scriptId), script, false)
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

	//
	// Run onLoad() function
	//
	tmr = time.AfterFunc(scriptLoadTimeout, func() {
		vm.Interrupt(errTimeout)
	})
	if fn, ok := goja.AssertFunction(vm.Get(`onLoad`)); ok {

		if _, err := fn(goja.Undefined(), vm.ToValue(mobActor)); err != nil {
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

	vmw := newVMWrapper(vm, 0)

	mobVMCache[scriptId] = vmw

	return vmw, nil
}
