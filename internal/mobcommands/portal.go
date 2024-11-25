package mobcommands

import (
	"fmt"
	"strconv"

	"github.com/volte6/gomud/internal/colorpatterns"
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/exit"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/rooms"
)

// Mob portaling is different than player portaling.
// Mob portals are open for shorter periods, and go to specific locations.
func Portal(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	// This is a hack because using "portal" to enter an existing portal is very common
	if rest == `` {
		if handled, err := Go(`portal`, mob, room); handled {
			return handled, err
		}
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
		if qty == 0 { // could't find any
			// No more rooms with items? Our job is done i guess.

			mob.Command(`portal home;drop all`)

			return true, fmt.Errorf("failed to find temporary exit to room")
		}
		portalTargetRoomId = mostItemRoomId
	}

	// Portal to home room
	if rest == `home` {
		portalTargetRoomId = mob.HomeRoomId
	}

	if portalTargetRoomId == mob.Character.RoomId {
		return true, fmt.Errorf("already in that room")
	}

	// Load current room details
	targetRoom := rooms.LoadRoom(portalTargetRoomId)
	if targetRoom == nil {
		return false, fmt.Errorf(`room %d not found`, portalTargetRoomId)
	}

	// Target = portalTargetRoomId
	// Current = user.Character.RoomId
	// At this point we have no open portals, we can create a new one.
	newPortalExitName := `dark portal`
	newPortal := exit.TemporaryRoomExit{
		RoomId:  portalTargetRoomId,
		Title:   colorpatterns.ApplyColorPattern(newPortalExitName, `gray`),
		UserId:  0,
		Expires: `2 rounds`,
	}

	// Spawn a portal in the room that leads to the portal location
	if !room.AddTemporaryExit(newPortalExitName, newPortal) {
		return true, fmt.Errorf("failed to add temporary exit to room")
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

	return true, nil
}
