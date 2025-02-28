// Round ticks for players
package hooks

import (
	"fmt"
	"time"

	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mobcommands"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/scripting"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

//
// Handle mobs that are bored
//

func IdleMobs(e events.Event) bool {

	evt := e.(events.NewRound)

	maxBoredom := uint8(evt.Config.MaxMobBoredom)
	globalConverseChance := int(evt.Config.MobConverseChance)

	allMobInstances := mobs.GetAllMobInstanceIds()

	allowedUnloadCt := len(allMobInstances) - int(evt.Config.MobUnloadThreshold)
	if allowedUnloadCt < 0 {
		allowedUnloadCt = 0
	}

	// Handle idle mob behavior
	tStart := time.Now()
	for _, mobId := range allMobInstances {

		mob := mobs.GetInstance(mobId)
		if mob == nil {
			allowedUnloadCt--
			continue
		}

		if allowedUnloadCt > 0 && mob.BoredomCounter >= maxBoredom {

			if mob.Despawns() {
				mob.Command(`despawn` + fmt.Sprintf(` depression %d/%d`, mob.BoredomCounter, maxBoredom))
				allowedUnloadCt--

			} else {
				mob.BoredomCounter = 0
			}

			continue
		}

		// If idle prevented, it's a one round interrupt (until another comes along)
		if mob.PreventIdle {
			mob.PreventIdle = false
			continue
		}

		// If they are doing some sort of combat thing,
		// Don't do idle actions
		if mob.Character.Aggro != nil {
			if mob.Character.Aggro.UserId > 0 {
				user := users.GetByUserId(mob.Character.Aggro.UserId)
				if user == nil || user.Character.RoomId != mob.Character.RoomId {
					mob.Command(`emote mumbles about losing their quarry.`)
					mob.Character.Aggro = nil
				}
			}
			continue
		}

		if mob.InConversation() {
			mob.Converse()
			continue
		}

		if mob.CanConverse() && util.Rand(100) < globalConverseChance {
			if mobRoom := rooms.LoadRoom(mob.Character.RoomId); mobRoom != nil {
				mobcommands.Converse(``, mob, mobRoom) // Execute this directly so that target mob doesn't leave the room before this command executes
				//mob.Command(`converse`)
			}
			continue
		}

		// If they have idle commands, maybe do one of them?
		handled, _ := scripting.TryMobScriptEvent("onIdle", mob.InstanceId, 0, ``, nil)
		if !handled {

			if !mob.Character.IsCharmed() { // Won't do this stuff if befriended

				if mob.MaxWander > -1 && len(mob.RoomStack) > mob.MaxWander {
					mob.GoingHome = true
				}

				if mob.GoingHome {
					mob.Command(`go home`)
					continue
				}

			}

			//
			// Look for trouble
			//
			if mob.Character.IsCharmed() {
				// Only some mobs can apply first aid
				if mob.Character.KnowsFirstAid() {
					mob.Command(`lookforaid`)
				}
			} else {

				idleCmd := `lookfortrouble`
				if util.Rand(100) < mob.ActivityLevel {
					idleCmd = mob.GetIdleCommand()
					if idleCmd == `` {
						idleCmd = `lookfortrouble`
					}
				}
				mob.Command(idleCmd)
			}
		}

	}

	util.TrackTime(`IdleMobs()`, time.Since(tStart).Seconds())

	return true
}
