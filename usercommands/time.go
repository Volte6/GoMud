package usercommands

import (
	"github.com/volte6/gomud/gametime"
	"github.com/volte6/gomud/rooms"
	"github.com/volte6/gomud/users"
)

func Time(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	gd := gametime.GetDate()

	dayNight := `day`
	if gd.Night {
		dayNight = `night`
	}

	user.SendText(`It is now ` + gd.String() + `. It is ` + dayNight + `.`)

	return true, nil
}
