package usercommands

import (
	"github.com/volte6/gomud/internal/characters"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
)

func Flee(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	if user.Character.Aggro == nil {
		user.SendText(`You aren't in combat!`)
	} else {
		user.SendText(`You attempt to flee...`)
		user.Character.Aggro.Type = characters.Flee
	}

	return true, nil
}