package usercommands

import (
	"fmt"

	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/gametime"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
)

func Time(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	gd := gametime.GetDate()

	if rest != `` { // testing code
		gd = gametime.GetDate(gametime.GetLastPeriod(rest, gd.RoundNumber))
	}

	dayNight := `day`
	if gd.Night {
		dayNight = `night`
	}

	user.SendText(fmt.Sprintf(`It is now %s. It is <ansi fg="%s">%stime</ansi> on <ansi fg="230">day %d</ansi> of <ansi fg="230">year %d</ansi>. The month is <ansi fg="230">%s</ansi>, and it is the year of the <ansi fg="230">%s</ansi>`,
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
