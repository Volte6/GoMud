package hooks

import (
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/users"
)

// Checks whether their level is too high for a guide
func CheckGuide(e events.Event) bool {

	evt, typeOk := e.(events.LevelUp)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "LevelUp", "Actual Type", e.Type())
		return false
	}

	user := users.GetByUserId(evt.UserId)
	if user == nil {
		return true
	}

	if user.Character.Level >= 5 {
		for _, mobInstanceId := range user.Character.CharmedMobs {
			if mob := mobs.GetInstance(mobInstanceId); mob != nil {

				if mob.MobId == 38 {
					mob.Command(`say I see you have grown much stronger and more experienced. My assistance is now needed elsewhere. I wish you good luck!`)
					mob.Command(`emote clicks their heels together and disappears in a cloud of smoke.`, 10)
					mob.Command(`suicide vanish`, 10)
				}
			}
		}
	}

	return true
}
