package scripting

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/dop251/goja"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

var (
	errNoScript = errors.New("no script")
	errTimeout  = errors.New("script timeout")
)

func Setup(scriptLoadTimeoutMs int, scriptRoomTimeoutMs int) {
	scriptLoadTimeout = time.Duration(scriptLoadTimeoutMs) * time.Millisecond
	scriptRoomTimeout = time.Duration(scriptRoomTimeoutMs) * time.Millisecond
}

func setAllScriptingFunctions(vm *goja.Runtime) {
	setMessagingFunctions(vm)
	setRoomFunctions(vm)
	setUserFunctions(vm)
	setUtilFunctions(vm)
	setMobFunctions(vm)
}

func TryScriptLoad(roomId int) error {
	timestart := time.Now()
	defer func() {
		slog.Debug("scripting/TryScriptLoad()", "time", time.Since(timestart))
	}()

	vmw, err := getRoomVM(roomId)
	if err != nil {
		return err
	}

	if onCommandFunc, ok := vmw.GetFunction(`onLoad`); ok {

		disableMessageQueue = true

		vmw.MarkUsed() // Mark the VM as used to prevent it from being pruned

		tmr := time.AfterFunc(scriptLoadTimeout, func() {
			vmw.VM.Interrupt(errTimeout)
		})
		_, err := onCommandFunc(goja.Undefined(),
			vmw.VM.ToValue(roomId),
		)
		vmw.VM.ClearInterrupt()
		tmr.Stop()

		disableMessageQueue = false

		if err != nil {

			// Wrap the error
			finalErr := fmt.Errorf("onLoad(): %w", err)

			if _, ok := finalErr.(*goja.Exception); ok {
				slog.Error("JSVM", "exception", finalErr)
				return finalErr
			} else if errors.Is(finalErr, errTimeout) {
				slog.Error("JSVM", "interrupted", finalErr)
				return finalErr
			}

			slog.Error("JSVM", "error", finalErr)
			return finalErr
		}

	}

	return nil
}

func TryScriptEvent(eventName string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {
	timestart := time.Now()
	defer func() {
		slog.Debug("scripting/TryScriptEvent()", "time", time.Since(timestart))
	}()

	messageQueue = util.NewMessageQueue(userId, 0)
	commandQueue = cmdQueue

	user := users.GetByUserId(userId)
	if user == nil {
		return messageQueue, errors.New("user not found")
	}

	vmw, err := getRoomVM(user.Character.RoomId)
	if err != nil {
		return messageQueue, err
	}

	if onCommandFunc, ok := vmw.GetFunction(eventName); ok {

		vmw.MarkUsed() // Mark the VM as used to prevent it from being pruned

		tmr := time.AfterFunc(scriptRoomTimeout, func() {
			vmw.VM.Interrupt(errTimeout)
		})
		res, err := onCommandFunc(goja.Undefined(),
			vmw.VM.ToValue(userId),
			vmw.VM.ToValue(user.Character.RoomId),
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

func TryCommand(cmd string, rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	timestart := time.Now()
	defer func() {
		slog.Debug("scripting/TryCommand()", "time", time.Since(timestart))
	}()

	messageQueue = util.NewMessageQueue(userId, 0)
	commandQueue = cmdQueue

	user := users.GetByUserId(userId)
	if user == nil {
		return messageQueue, errors.New("user not found")
	}

	vmw, err := getRoomVM(user.Character.RoomId)
	if err != nil {
		return messageQueue, err
	}

	if onCommandFunc, ok := vmw.GetFunction(`onCommand_` + cmd); ok {

		slog.Info("onCommandFunc", "FOUND", `onCommand_`+cmd)

		vmw.MarkUsed() // Mark the VM as used to prevent it from being pruned

		tmr := time.AfterFunc(scriptRoomTimeout, func() {
			vmw.VM.Interrupt(errTimeout)
		})
		res, err := onCommandFunc(goja.Undefined(),
			vmw.VM.ToValue(rest),
			vmw.VM.ToValue(userId),
			vmw.VM.ToValue(user.Character.RoomId),
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

		vmw.MarkUsed() // Mark the VM as used to prevent it from being pruned

		tmr := time.AfterFunc(scriptRoomTimeout, func() {
			vmw.VM.Interrupt(errTimeout)
		})
		res, err := onCommandFunc(goja.Undefined(),
			vmw.VM.ToValue(cmd),
			vmw.VM.ToValue(rest),
			vmw.VM.ToValue(userId),
			vmw.VM.ToValue(user.Character.RoomId),
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
