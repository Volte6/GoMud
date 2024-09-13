package mobcommands

import (
	"fmt"
	"strconv"
	"time"

	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
)

// Mob portaling is different than player portaling.
// Mob portals are open for shorter periods, and go to specific locations.
func Portal(rest string, mobId int) (bool, string, error) {

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return false, ``, fmt.Errorf("mob %d not found", mobId)
	}

	// This is a hack because using "portal" to enter an existing portal is very common
	if rest == `` {
		if handled, nextCommand, err := Go(`portal`, mobId); handled {
			return handled, nextCommand, err
		}
	}

	// Load current room details
	room := rooms.LoadRoom(mob.Character.RoomId)
	if room == nil {
		return false, ``, fmt.Errorf(`room %d not found`, mob.Character.RoomId)
	}

	var err error

	// Establish the default portal location
	portalTargetRoomId := -1 // Default to their own home

	// Portal to a specific room?
	if rest != `` {
		portalTargetRoomId, err = strconv.Atoi(rest)
		if err != nil {
			portalTargetRoomId = mob.HomeRoomId
		}
		err = nil
	}

	// Portal to the highest loot room
	if rest == `loot` {

		config := configs.GetConfig()

		// Only interest in rooms where players haven't visited in a while and have at least 1
		mostItemRoomId, qty := rooms.GetRoomWithMostItems(bool(config.LootGoblinIncludeRecentRooms), int(config.LootGoblinMinimumItems), int(config.LootGoblinMinimumGold))
		if portalTargetRoomId == 0 && qty == 0 { // could't find any
			// No more rooms with items? Our job is done i guess.

			mob.Command(`portal home;drop all`)

			return true, ``, fmt.Errorf("failed to find temporary exit to room")
		}
		portalTargetRoomId = mostItemRoomId
	}

	// Portal to home room
	if rest == `home` {
		portalTargetRoomId = mob.HomeRoomId
	}

	if portalTargetRoomId == mob.Character.RoomId {
		return false, ``, err
	}

	// Load current room details
	targetRoom := rooms.LoadRoom(portalTargetRoomId)
	if targetRoom == nil {
		return false, ``, fmt.Errorf(`room %d not found`, portalTargetRoomId)
	}

	// Target = portalTargetRoomId
	// Current = user.Character.RoomId
	// At this point we have no open portals, we can create a new one.
	newPortalExitName := `dark portal`
	newPortal := rooms.TemporaryRoomExit{
		RoomId:  portalTargetRoomId,
		Title:   fmt.Sprintf(`<ansi fg="black-bold">%s</ansi>`, newPortalExitName),
		UserId:  0,
		Expires: time.Now().Add(time.Duration(configs.GetConfig().RoundSeconds*2) * time.Second), // lasts for 2 rounds
	}

	// Spawn a portal in the room that leads to the portal location
	if !room.AddTemporaryExit(newPortalExitName, newPortal) {
		return true, ``, fmt.Errorf("failed to add temporary exit to room")
	}

	room.SendText(
		fmt.Sprintf(`<ansi fg="mobname">%s</ansi> squints really hard, and a %s appears!`, mob.Character.Name, newPortal.Title),
	)

	// Modify it for this room
	newPortal.RoomId = mob.Character.RoomId
	targetRoom.AddTemporaryExit(newPortalExitName, newPortal)

	room.SendText(
		fmt.Sprintf(`A %s appears!`, newPortal.Title),
	)

	return true, ``, nil
}
