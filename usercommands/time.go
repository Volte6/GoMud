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

	if rest != `` {
		gd = gametime.GetDate(gd.AddPeriod(rest))
	}

	user.SendText(fmt.Sprintf(`It is now %s. It is <ansi fg="%s">%stime</ansi> on <ansi fg="230">day %d</ansi> of the <ansi fg="230">year %d</ansi>. The month is <ansi fg="230">%s</ansi>, and it is the year of the <ansi fg="230">%s</ansi>`,
		gd.String(),
		dayNight,
		dayNight,
		gd.Day,
		gd.Year,
		gametime.MonthName(gd.Month),
		gametime.GetZodiac(gd.Year),
	))

	return true, nil
}
