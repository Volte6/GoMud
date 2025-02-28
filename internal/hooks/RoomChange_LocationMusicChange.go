package hooks

import (
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
)

//
// RoomChangeHandler waits for RoomChange events
// It then sends out GMCP data updates
// Also sends music changes out
//

func LocationMusicChange(e events.Event) bool {

	evt, typeOk := e.(events.RoomChange)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "RoomChange", "Actual Type", e.Type())
		return false
	}

	// If this isn't a user changing rooms, just pass it along.
	if evt.UserId == 0 {
		return true
	}

	// Get user... Make sure they still exist too.
	user := users.GetByUserId(evt.UserId)
	if user == nil {
		return false
	}

	// Get the new room data... abort if doesn't exist.
	newRoom := rooms.LoadRoom(evt.ToRoomId)
	if newRoom == nil {
		return false
	}

	// Get the old room data... abort if doesn't exist.
	oldRoom := rooms.LoadRoom(evt.FromRoomId)
	if oldRoom == nil {
		return false
	}

	// If this zone has music, play it.
	// Room music takes priority.
	if newRoom.MusicFile != `` {
		user.PlayMusic(newRoom.MusicFile)
	} else {
		zoneInfo := rooms.GetZoneConfig(newRoom.Zone)
		if zoneInfo.MusicFile != `` {
			user.PlayMusic(zoneInfo.MusicFile)
		} else if oldRoom.MusicFile != `` {
			user.PlayMusic(`Off`)
		}
	}

	return true
}
