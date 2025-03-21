package modules

import (
	"embed"
	"fmt"

	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/gametime"
	"github.com/volte6/gomud/internal/plugins"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
)

var (

	//////////////////////////////////////////////////////////////////////
	// NOTE: The below //go:embed directive is important!
	// It embeds the relative path into the var below it.
	//////////////////////////////////////////////////////////////////////

	//go:embed time/*
	time_Files embed.FS // All vars must be a unique name since the module package/namespace is shared between modules.
)

// ////////////////////////////////////////////////////////////////////
// NOTE: The init function in Go is a special function that is
// automatically executed before the main function within a package.
// It is used to initialize variables, set up configurations, or
// perform any other setup tasks that need to be done before the
// program starts running.
// ////////////////////////////////////////////////////////////////////
func init() {

	//
	// We can use all functions only, but this demonstrates
	//
	plug := plugins.New(`time`, `1.0`)

	//
	// Add the embedded filesystem
	//
	if err := plug.AttachFileSystem(time_Files); err != nil {
		panic(err)
	}
	//
	// Register any user/mob commands
	//
	plug.AddUserCommand(`time`, TimeCommand, true, false)
}

//////////////////////////////////////////////////////////////////////
// NOTE: What follows is all custom code. For this module.
//////////////////////////////////////////////////////////////////////

func TimeCommand(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

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
