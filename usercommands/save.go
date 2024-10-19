package usercommands

import (
	"github.com/volte6/gomud/rooms"
	"github.com/volte6/gomud/users"
)

func Save(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	user.SendText("Saving...")
	users.SaveUser(*user)
	user.SendText("done.")

	return true, nil
}
