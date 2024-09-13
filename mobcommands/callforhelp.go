package mobcommands

import (
	"fmt"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/util"
)

// Should check adjacent rooms for mobs and call them into the room to help if of the same group
// Format should be:
// callforhelp blows his horn
// "blows his horn" will be emoted to the room
func CallForHelp(rest string, mobId int) (util.MessageQueue, error) {

	response := NewMobCommandResponse(mobId)

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("mob %d not found", mobId)
	}

	// Load current room details
	room := rooms.LoadRoom(mob.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, mob.Character.RoomId)
	}

	if mob.Character.Aggro == nil || mob.Character.Aggro.UserId == 0 {
		return response, fmt.Errorf(`mob %d has no aggro`, mobId)
	}

	calledForHelp := false

	for _, roomInfo := range room.Exits {
		adjRoom := rooms.LoadRoom(roomInfo.RoomId)
		if adjRoom.MobCt() < 1 {
			continue
		}

		exitIntoRoom := adjRoom.FindExitTo(room.RoomId)
		if exitIntoRoom == `` {
			continue
		}

		for _, nearbyMobInstanceId := range adjRoom.GetMobs(rooms.FindNeutral, rooms.FindHostile) {
			if mobInfo := mobs.GetInstance(nearbyMobInstanceId); mobInfo != nil {

				//if mobInfo.MaxWander == 0 { // Mobs that do not wander at all won't heed the call
				//	continue
				//}

				if !mobInfo.IsAlly(mob) { // Only help allies
					continue
				}

				if !calledForHelp {
					calledForHelp = true

					if rest != `` {
						mob.Command(fmt.Sprintf(`emote %s`, rest))
					} else {
						mob.Command(`emote calls for help`)
					}
				}

				mobInfo.Command(fmt.Sprintf(`go %s`, exitIntoRoom))
				mobInfo.Command(fmt.Sprintf(`attack @%d`, mob.Character.Aggro.UserId))
			}
		}
	}

	response.Handled = true
	return response, nil
}
