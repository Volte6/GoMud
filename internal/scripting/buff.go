package scripting

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/dop251/goja"
	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/colorpatterns"
)

var (
	buffVMCache       = make(map[int]*VMWrapper)
	scriptBuffTimeout = 50 * time.Millisecond
)

func ClearBuffVMs() {
	clear(buffVMCache)
}

func PruneBuffVMs(instanceIds ...int) {
	// Do not prune, they dont' get a VM per buff instance.
}

func TryBuffScriptEvent(eventName string, userId int, mobInstanceId int, buffId int) (bool, error) {

	slog.Info("TryBuffScriptEvent()", "eventName", eventName, "buffId", buffId)
	vmw, err := getBuffVM(buffId)
	if err != nil {
		return false, err
	}

	actorInfo := GetActor(userId, mobInstanceId)
	buffTriggersLeft := actorInfo.characterRecord.Buffs.TriggersLeft(buffId)

	timestart := time.Now()
	defer func() {
		slog.Debug("TryBuffScriptEvent()", "eventName", eventName, "buffId", buffId, "time", time.Since(timestart))
	}()
	if onCommandFunc, ok := vmw.GetFunction(eventName); ok {

		// Set forced ansi tag wrappers
		userTextWrap.Set(`buff-text`, ``, `cyan`, colorpatterns.Stretch)
		roomTextWrap.Set(`buff-text`, ``, `cyan`, colorpatterns.Stretch)

		tmr := time.AfterFunc(scriptRoomTimeout, func() {
			vmw.VM.Interrupt(errTimeout)
		})

		res, err := onCommandFunc(goja.Undefined(),
			vmw.VM.ToValue(actorInfo),
			vmw.VM.ToValue(buffTriggersLeft),
		)
		vmw.VM.ClearInterrupt()
		tmr.Stop()

		// Reset forced ansi tag wrappers
		userTextWrap.Reset()
		roomTextWrap.Reset()

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

func TryBuffCommand(cmd string, rest string, userId int, mobInstanceId int, buffId int) (bool, error) {

	vmw, err := getBuffVM(buffId)
	if err != nil {
		return false, err
	}

	sActor := GetActor(userId, mobInstanceId)
	sRoom := GetRoom(sActor.GetRoomId())

	timestart := time.Now()
	defer func() {
		slog.Debug("TryBuffCommand()", "cmd", cmd, "buffId", buffId, "time", time.Since(timestart))
	}()

	if onCommandFunc, ok := vmw.GetFunction(`onCommand_` + cmd); ok {

		tmr := time.AfterFunc(scriptRoomTimeout, func() {
			vmw.VM.Interrupt(errTimeout)
		})
		res, err := onCommandFunc(goja.Undefined(),
			vmw.VM.ToValue(rest),
			vmw.VM.ToValue(sActor),
			vmw.VM.ToValue(sRoom),
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

		sActor := GetActor(userId, mobInstanceId)
		sRoom := GetRoom(sActor.GetRoomId())

		tmr := time.AfterFunc(scriptRoomTimeout, func() {
			vmw.VM.Interrupt(errTimeout)
		})
		res, err := onCommandFunc(goja.Undefined(),
			vmw.VM.ToValue(cmd),
			vmw.VM.ToValue(rest),
			vmw.VM.ToValue(sActor),
			vmw.VM.ToValue(sRoom),
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

func getBuffVM(buffId int) (*VMWrapper, error) {

	if vm, ok := buffVMCache[buffId]; ok {
		if vm == nil {
			return nil, errNoScript
		}
		return vm, nil
	}

	bSpec := buffs.GetBuffSpec(buffId)
	if bSpec == nil {
		return nil, fmt.Errorf("buff spec not found: %d", bSpec)
	}

	script := bSpec.GetScript()
	if len(script) == 0 {
		buffVMCache[buffId] = nil
		return nil, errNoScript
	}

	vm := goja.New()
	setAllScriptingFunctions(vm)

	prg, err := goja.Compile(fmt.Sprintf(`buff-%d`, buffId), script, false)
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

	buffVMCache[buffId] = vmw

	return vmw, nil
}
