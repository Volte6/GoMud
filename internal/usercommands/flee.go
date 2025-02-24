package usercommands

import (
	"github.com/volte6/gomud/internal/characters"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
)

func Flee(rest string, user *users.UserRecord, room *rooms.Room, flags UserCommandFlag) (bool, error) {

	if user.Character.Aggro == nil || user.Character.Aggro.Type != characters.Flee {
		user.SendText(`You attempt to flee...`)

		user.Character.Aggro = &characters.Aggro{}
		user.Character.Aggro.Type = characters.Flee
	}

	return true, nil
}
