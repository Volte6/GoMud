package hooks

import (
	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/characters"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/scripting"
	"github.com/volte6/gomud/internal/users"
)

//
// Checks for quests on the item
//

func ApplyBuffs(e events.Event) events.EventReturn {

	evt, typeOk := e.(events.Buff)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "Buff", "Actual Type", e.Type())
		return events.Cancel
	}

	//mudlog.Debug(`Event`, `type`, evt.Type(), `UserId`, evt.UserId, `MobInstanceId`, evt.MobInstanceId, `BuffId`, evt.BuffId)

	buffInfo := buffs.GetBuffSpec(evt.BuffId)
	if buffInfo == nil {
		return events.Cancel
	}

	var targetChar *characters.Character

	if evt.MobInstanceId > 0 {

		buffMob := mobs.GetInstance(evt.MobInstanceId)
		if buffMob == nil {
			return events.Cancel
		}

		targetChar = &buffMob.Character

	} else {

		buffUser := users.GetByUserId(evt.UserId)
		if buffUser == nil {
			return events.Cancel
		}

		targetChar = buffUser.Character
	}

	if evt.BuffId < 0 {
		targetChar.RemoveBuff(buffInfo.BuffId * -1)
		return events.Continue
	}

	// Apply the buff
	targetChar.AddBuff(evt.BuffId, false)

	//
	// Fire onStart for buff script
	//
	if _, err := scripting.TryBuffScriptEvent(`onStart`, evt.UserId, evt.MobInstanceId, evt.BuffId); err == nil {
		targetChar.TrackBuffStarted(evt.BuffId)
	}

	//
	// If the buff calls for an immediate triggering
	//
	if buffInfo.TriggerNow {
		scripting.TryBuffScriptEvent(`onTrigger`, evt.UserId, evt.MobInstanceId, evt.BuffId)

		if evt.MobInstanceId > 0 && targetChar.Health <= 0 {
			// Mob died
			events.AddToQueue(events.Input{
				MobInstanceId: evt.MobInstanceId,
				InputText:     `suicide`,
			})
		}
	}

	return events.Continue
}
