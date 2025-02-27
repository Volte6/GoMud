package scripting

import (
	"errors"
	"fmt"
	"time"

	"github.com/dop251/goja"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
)

var (
	roomVMCache       = make(map[int]*VMWrapper)
	scriptLoadTimeout = 1000 * time.Millisecond
	scriptRoomTimeout = 50 * time.Millisecond
)

func ClearRoomVMs() {
	clear(roomVMCache)
}

func PruneRoomVMs(roomIds ...int) {
	if len(roomIds) > 0 {
		for _, roomId := range roomIds {
			delete(roomVMCache, roomId)
		}
		return
	}
	for roomId, _ := range roomVMCache {
		if !rooms.IsRoomLoaded(roomId) {
			delete(roomVMCache, roomId)
		}
	}
}

func TryRoomScriptEvent(eventName string, userId int, roomId int) (bool, error) {

	vmw, err := getRoomVM(roomId)
	if err != nil {
		return false, err
	}

	timestart := time.Now()
	defer func() {
		mudlog.Debug("TryRoomScriptEvent()", "eventName", eventName, "roomId", roomId, "time", time.Since(timestart))
	}()

	if onCommandFunc, ok := vmw.GetFunction(eventName); ok {

		// Set forced ansi tag wrappers
		userTextWrap.Set(`script-text`, ``, ``)
		roomTextWrap.Set(`script-text`, ``, ``)

		sUser := GetActor(userId, 0)
		sRoom := GetRoom(roomId)

		tmr := time.AfterFunc(scriptRoomTimeout, func() {
			vmw.VM.Interrupt(errTimeout)
		})

		res, err := onCommandFunc(goja.Undefined(),
			vmw.VM.ToValue(sUser),
			vmw.VM.ToValue(sRoom),
		)

		vmw.VM.ClearInterrupt()
		tmr.Stop()

		userTextWrap.Reset()
		roomTextWrap.Reset()

		if err != nil {

			// Wrap the error
			finalErr := fmt.Errorf("%s(): %w", eventName, err)

			if _, ok := finalErr.(*goja.Exception); ok {
				mudlog.Error("JSVM", "exception", finalErr)
				return false, finalErr
			} else if errors.Is(finalErr, errTimeout) {
				mudlog.Error("JSVM", "interrupted", finalErr)
				return false, finalErr
			}

			mudlog.Error("JSVM", "error", finalErr)
			return false, finalErr
		}

		if boolVal, ok := res.Export().(bool); ok {
			return boolVal, nil
		}
	}

	return false, nil
}

func TryRoomIdleEvent(roomId int) (bool, error) {

	vmw, err := getRoomVM(roomId)
	if err != nil {
		return false, err
	}

	timestart := time.Now()
	defer func() {
		mudlog.Debug("TryRoomIdleEvent()", "roomId", roomId, "time", time.Since(timestart))
	}()

	if onCommandFunc, ok := vmw.GetFunction(`onIdle`); ok {

		// Set forced ansi tag wrappers
		userTextWrap.Set(`script-text`, ``, ``)
		roomTextWrap.Set(`script-text`, ``, ``)

		sRoom := GetRoom(roomId)

		tmr := time.AfterFunc(scriptRoomTimeout, func() {
			vmw.VM.Interrupt(errTimeout)
		})

		res, err := onCommandFunc(goja.Undefined(),
			vmw.VM.ToValue(sRoom),
		)

		vmw.VM.ClearInterrupt()
		tmr.Stop()

		userTextWrap.Reset()
		roomTextWrap.Reset()

		if err != nil {

			// Wrap the error
			finalErr := fmt.Errorf("TryRoomIdleEvent(): %w", err)

			if _, ok := finalErr.(*goja.Exception); ok {
				mudlog.Error("JSVM", "exception", finalErr)
				return false, finalErr
			} else if errors.Is(finalErr, errTimeout) {
				mudlog.Error("JSVM", "interrupted", finalErr)
				return false, finalErr
			}

			mudlog.Error("JSVM", "error", finalErr)
			return false, finalErr
		}

		if boolVal, ok := res.Export().(bool); ok {
			return boolVal, nil
		}
	}

	return false, nil
}

func TryRoomCommand(cmd string, rest string, userId int) (bool, error) {

	user := users.GetByUserId(userId)
	if user == nil {
		return false, errors.New("user not found")
	}

	room := rooms.LoadRoom(user.Character.RoomId)

	altCmd, _ := room.FindExitByName(cmd)

	if room != nil {

		for _, mobInstanceId := range room.GetMobs() {
			if handled, err := TryMobCommand(cmd, rest, mobInstanceId, userId, `user`); err == nil {
				if handled {
					return true, nil
				}
			}

		}
	}

	vmw, err := getRoomVM(user.Character.RoomId)
	if err != nil {
		return false, err
	}

	timestart := time.Now()
	defer func() {
		mudlog.Debug("TryRoomCommand()", "cmd", cmd, "roomId", user.Character.RoomId, "time", time.Since(timestart))
	}()

	onCommandFunc, cmdFound := vmw.GetFunction(`onCommand_` + cmd)
	if !cmdFound && altCmd != `` {
		onCommandFunc, cmdFound = vmw.GetFunction(`onCommand_` + altCmd)
	}

	if cmdFound {

		// Set forced ansi tag wrappers
		userTextWrap.Set(`script-text`, ``, ``)
		roomTextWrap.Set(`script-text`, ``, ``)

		sUser := GetUser(userId)
		sRoom := GetRoom(user.Character.RoomId)

		tmr := time.AfterFunc(scriptRoomTimeout, func() {
			vmw.VM.Interrupt(errTimeout)
		})
		res, err := onCommandFunc(goja.Undefined(),
			vmw.VM.ToValue(rest),
			vmw.VM.ToValue(sUser),
			vmw.VM.ToValue(sRoom),
		)
		vmw.VM.ClearInterrupt()
		tmr.Stop()

		userTextWrap.Reset()
		roomTextWrap.Reset()

		if err != nil {

			// Wrap the error
			finalErr := fmt.Errorf("onCommand_%s(): %w", cmd, err)

			if _, ok := finalErr.(*goja.Exception); ok {
				mudlog.Error("JSVM", "exception", finalErr)
				return false, finalErr
			} else if errors.Is(finalErr, errTimeout) {
				mudlog.Error("JSVM", "interrupted", finalErr)
				return false, finalErr
			}

			mudlog.Error("JSVM", "error", finalErr)
			return false, finalErr
		}

		if boolVal, ok := res.Export().(bool); ok {
			return boolVal, nil
		}

	} else if onCommandFunc, ok := vmw.GetFunction(`onCommand`); ok {

		// Set forced ansi tag wrappers
		userTextWrap.Set(`script-text`, ``, ``)
		roomTextWrap.Set(`script-text`, ``, ``)

		sUser := GetUser(userId)
		sRoom := GetRoom(user.Character.RoomId)

		tmr := time.AfterFunc(scriptRoomTimeout, func() {
			vmw.VM.Interrupt(errTimeout)
		})
		res, err := onCommandFunc(goja.Undefined(),
			vmw.VM.ToValue(cmd),
			vmw.VM.ToValue(rest),
			vmw.VM.ToValue(sUser),
			vmw.VM.ToValue(sRoom),
		)
		vmw.VM.ClearInterrupt()
		tmr.Stop()

		userTextWrap.Reset()
		roomTextWrap.Reset()

		if err != nil {

			// Wrap the error
			finalErr := fmt.Errorf("onCommand(): %w", err)

			if _, ok := finalErr.(*goja.Exception); ok {
				mudlog.Error("JSVM", "exception", finalErr)
				return false, finalErr
			} else if errors.Is(finalErr, errTimeout) {
				mudlog.Error("JSVM", "interrupted", finalErr)
				return false, finalErr
			}

			mudlog.Error("JSVM", "error", finalErr)
			return false, finalErr
		}

		if boolVal, ok := res.Export().(bool); ok {
			return boolVal, nil
		}
	}

	return false, nil
}

func getRoomVM(roomId int) (*VMWrapper, error) {

	if vm, ok := roomVMCache[roomId]; ok {
		if vm == nil {
			return nil, errNoScript
		}
		return vm, nil
	}

	room := rooms.LoadRoom(roomId)
	if room == nil {
		return nil, fmt.Errorf("room not found: %d", roomId)
	}

	script := room.GetScript()
	if len(script) == 0 {
		roomVMCache[roomId] = nil
		return nil, errNoScript
	}

	vm := goja.New()
	setAllScriptingFunctions(vm)

	prg, err := goja.Compile(fmt.Sprintf(`room-%d`, roomId), script, false)
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
			mudlog.Error("JSVM", "exception", finalErr)
			return nil, finalErr
		} else if errors.Is(finalErr, errTimeout) {
			mudlog.Error("JSVM", "interrupted", finalErr)
			return nil, finalErr
		}

		mudlog.Error("JSVM", "error", finalErr)
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

		sRoom := GetRoom(roomId)

		if _, err := fn(goja.Undefined(), vm.ToValue(sRoom)); err != nil {
			// Wrap the error
			finalErr := fmt.Errorf("onLoad: %w", err)

			if _, ok := finalErr.(*goja.Exception); ok {
				mudlog.Error("JSVM", "exception", finalErr)
				return nil, finalErr
			} else if errors.Is(finalErr, errTimeout) {
				mudlog.Error("JSVM", "interrupted", finalErr)
				return nil, finalErr
			}

			mudlog.Error("JSVM", "error", finalErr)
			return nil, finalErr
		}
	}
	vm.ClearInterrupt()
	tmr.Stop()

	vmw := newVMWrapper(vm, 0)

	roomVMCache[roomId] = vmw

	return vmw, nil
}
