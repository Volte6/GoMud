package usercommands

import (
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
)

// Default is a special command that tries to contextually pick a default action for a room.
// The failover is to "look"
func Default(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	// If there is a shop, "list"
	if len(room.GetMobs(rooms.FindMerchant)) > 0 || len(room.GetPlayers(rooms.FindMerchant)) > 0 {
		List(``, user, room)
		return true, nil
	}

	// If there is a trainer, "train"
	if len(room.SkillTraining) > 0 {
		Train(``, user, room)
		return true, nil
	}

	// If a bank, "bank"
	if room.IsBank {
		Bank(``, user, room)
		return true, nil
	}

	// If a storage location, "storage"
	if room.IsStorage {
		Storage(``, user, room)
		return true, nil
	}

	// Default to "look"
	Look(``, user, room)

	return true, nil
}
