package scripting

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/dop251/goja"
	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/spells"
	"github.com/volte6/mud/util"
)

var (
	spellVMCache       = make(map[string]*VMWrapper)
	scriptSpellTimeout = 10 * time.Millisecond
)

func PruneSpellVMs(instanceIds ...int) {

}

func TrySpellScriptEvent(eventName string, sourceUserId int, sourceMobInstanceId int, spellAggro characters.SpellAggroInfo, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	messageQueue = util.NewMessageQueue(sourceUserId, sourceMobInstanceId)
	commandQueue = cmdQueue

	spellInfo := spells.GetSpell(spellAggro.SpellId)
	if spellInfo == nil {
		return messageQueue, fmt.Errorf("spell %s not found", spellAggro.SpellId)
	}

	timestart := time.Now()
	defer func() {
		slog.Debug("TrySpellScriptEvent()", "eventName", eventName, "spellId", spellAggro.SpellId, "spellRest", spellAggro.SpellRest, "TargetUsers", spellAggro.TargetUserIds, "TargetMobs", spellAggro.TargetMobInstanceIds, "time", time.Since(timestart))
	}()

	vmw, err := getSpellVM(spellAggro.SpellId)
	if err != nil {
		slog.Debug("TrySpellScriptEvent()", "error", err)
		return messageQueue, err
	}

	sourceActor := GetActor(sourceUserId, sourceMobInstanceId)

	if eventName != `onCast` && eventName != `onWait` && eventName != `onMagic` && eventName != `onFail` {
		return messageQueue, err
	}

	var stringArg string = ""
	var singleTargetArg *ScriptActor = nil
	var multiTargetArg []*ScriptActor = nil

	if spellInfo.Type == spells.Neutral {

		// arg is just whatever the user entered after the spell casting command
		stringArg = spellAggro.SpellRest

	} else if spellInfo.Type == spells.HelpSingle || spellInfo.Type == spells.HarmSingle {

		// arg is a single actor
		if len(spellAggro.TargetUserIds) > 0 {
			singleTargetArg = GetActor(spellAggro.TargetUserIds[0], 0)
		} else if len(spellAggro.TargetMobInstanceIds) > 0 {
			singleTargetArg = GetActor(0, spellAggro.TargetMobInstanceIds[0])
		}

		// If no longer in the same room, notify the user
		if singleTargetArg == nil || (sourceActor.GetRoomId() != singleTargetArg.GetRoomId()) {
			messageQueue.SendUserMessage(sourceUserId, `Your target cannot be found.`, true)
			messageQueue.Handled = true
			return messageQueue, nil
		}

	} else if spellInfo.Type == spells.HelpMulti || spellInfo.Type == spells.HarmMulti {

		// arg is a list of actors
		multiTargetArg = []*ScriptActor{}
		for _, targetUserId := range spellAggro.TargetUserIds {
			if uActor := GetActor(targetUserId, 0); uActor != nil {
				if uActor.GetRoomId() == sourceActor.GetRoomId() {
					multiTargetArg = append(multiTargetArg, uActor)
				}
			}
		}
		for _, targetMobInstanceId := range spellAggro.TargetMobInstanceIds {
			if mActor := GetActor(0, targetMobInstanceId); mActor != nil {
				if mActor.GetRoomId() == sourceActor.GetRoomId() {
					multiTargetArg = append(multiTargetArg, mActor)
				}
			}
		}

		if len(multiTargetArg) == 0 {
			messageQueue.SendUserMessage(sourceUserId, `Your target cannot be found.`, true)
			messageQueue.Handled = true
			return messageQueue, nil
		}

	}

	if onCommandFunc, ok := vmw.GetFunction(eventName); ok {

		var argValue goja.Value
		if multiTargetArg != nil {
			argValue = vmw.VM.ToValue(multiTargetArg)
		} else if singleTargetArg != nil {
			argValue = vmw.VM.ToValue(singleTargetArg)
		} else {
			argValue = vmw.VM.ToValue(stringArg)
		}

		tmr := time.AfterFunc(scriptItemTimeout, func() {
			vmw.VM.Interrupt(errTimeout)
		})
		res, err := onCommandFunc(goja.Undefined(),
			vmw.VM.ToValue(sourceActor),
			vmw.VM.ToValue(argValue),
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

func getSpellVM(scriptId string) (*VMWrapper, error) {

	if vm, ok := itemVMCache[scriptId]; ok {
		if vm == nil {
			return nil, errNoScript
		}
		return vm, nil
	}

	spellData := spells.GetSpell(scriptId)
	if spellData == nil {
		return nil, fmt.Errorf("spell %s not found", scriptId)
	}

	script := spellData.GetScript()
	if len(script) == 0 {
		itemVMCache[scriptId] = nil
		return nil, errNoScript
	}

	vm := goja.New()
	setAllScriptingFunctions(vm)

	prg, err := goja.Compile(fmt.Sprintf(`spell-%s`, scriptId), script, false)
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

	itemVMCache[scriptId] = vmw

	return vmw, nil
}
