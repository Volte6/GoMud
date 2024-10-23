package usercommands

import (
	"fmt"

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

	gd = gd.AdjustTo(`hour`, 0, 0, 0)

	user.SendText(fmt.Sprintf(`It is now %s. It is <ansi fg="230">%s</ansi> on <ansi fg="230">day %d</ansi> of <ansi fg="230">year %d</ansi>.`,
		gd.String(),
		dayNight,
		gd.Day,
		gd.Year,
	))

	return true, nil
}
