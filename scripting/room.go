package scripting

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/dop251/goja"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

var (
	roomVMCache       = make(map[int]*VMWrapper)
	scriptLoadTimeout = 1000 * time.Millisecond
	scriptRoomTimeout = 10 * time.Millisecond
)

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

func TryRoomScriptEvent(eventName string, userId int, roomId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	messageQueue = util.NewMessageQueue(userId, 0)
	commandQueue = cmdQueue

	vmw, err := getRoomVM(roomId)
	if err != nil {
		return messageQueue, err
	}

	timestart := time.Now()
	defer func() {
		slog.Debug("TryRoomScriptEvent()", "eventName", eventName, "roomId", roomId, "time", time.Since(timestart))
	}()

	if onCommandFunc, ok := vmw.GetFunction(eventName); ok {

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

func TryRoomIdleEvent(roomId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	messageQueue = util.NewMessageQueue(0, 0)
	commandQueue = cmdQueue

	vmw, err := getRoomVM(roomId)
	if err != nil {
		return messageQueue, err
	}

	timestart := time.Now()
	defer func() {
		slog.Debug("TryRoomIdleEvent()", "roomId", roomId, "time", time.Since(timestart))
	}()

	if onCommandFunc, ok := vmw.GetFunction(`onIdle`); ok {

		sRoom := GetRoom(roomId)

		tmr := time.AfterFunc(scriptRoomTimeout, func() {
			vmw.VM.Interrupt(errTimeout)
		})

		res, err := onCommandFunc(goja.Undefined(),
			vmw.VM.ToValue(sRoom),
		)

		vmw.VM.ClearInterrupt()
		tmr.Stop()

		if err != nil {

			// Wrap the error
			finalErr := fmt.Errorf("TryRoomIdleEvent(): %w", err)

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

func TryRoomCommand(cmd string, rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	messageQueue = util.NewMessageQueue(userId, 0)
	commandQueue = cmdQueue

	user := users.GetByUserId(userId)
	if user == nil {
		return messageQueue, errors.New("user not found")
	}

	room := rooms.LoadRoom(user.Character.RoomId)
	if room != nil {

		/*
			// NOT SURE if I want to do this yet...
			// This would allow buffs to capture commands

			for _, buffInfo := range user.Character.GetBuffs() {
				if mq, err := TryBuffCommand(cmd, rest, userId, 0, buffInfo.BuffId, cmdQueue); err == nil {
					messageQueue.AbsorbMessages(mq)

					messageQueue.Handled = messageQueue.Handled || mq.Handled
					if messageQueue.Handled {
						return messageQueue, nil
					}
				}
			}
		*/

		for _, mobInstanceId := range room.GetMobs() {
			if mq, err := TryMobCommand(cmd, rest, mobInstanceId, userId, `user`, cmdQueue); err == nil {
				messageQueue.AbsorbMessages(mq)

				messageQueue.Handled = messageQueue.Handled || mq.Handled
				if messageQueue.Handled {
					return messageQueue, nil
				}
			}

		}
	}

	vmw, err := getRoomVM(user.Character.RoomId)
	if err != nil {
		return messageQueue, err
	}

	timestart := time.Now()
	defer func() {
		slog.Debug("TryRoomCommand()", "cmd", cmd, "roomId", user.Character.RoomId, "time", time.Since(timestart))
	}()

	if onCommandFunc, ok := vmw.GetFunction(`onCommand_` + cmd); ok {

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

		sRoom := GetRoom(roomId)

		if _, err := fn(goja.Undefined(), vm.ToValue(sRoom)); err != nil {
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

	roomVMCache[roomId] = vmw

	return vmw, nil
}
