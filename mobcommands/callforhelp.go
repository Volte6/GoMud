package mobcommands

import (
	"fmt"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
)

// Should check adjacent rooms for mobs and call them into the room to help if of the same group
// Format should be:
// callforhelp blows his horn
// "blows his horn" will be emoted to the room
func CallForHelp(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	if mob.Character.Aggro == nil || mob.Character.Aggro.UserId == 0 {
		return false, fmt.Errorf(`mob %d has no aggro`, mob.InstanceId)
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

	return true, nil
}
