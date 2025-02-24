package usercommands

import (
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
)

func Save(rest string, user *users.UserRecord, room *rooms.Room, flags UserCommandFlag) (bool, error) {

	user.SendText("Saving...")
	users.SaveUser(*user)
	user.SendText("done.")

	return true, nil
}
