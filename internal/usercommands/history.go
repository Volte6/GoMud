package usercommands

import (
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
)

func History(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	headers := []string{`Type` /*`Round`,*/, `Time`, `Log`}

	rows := [][]string{}

	formatting := []string{
		`<ansi fg="red">%s</ansi>`,
		//`<ansi fg="red">%s</ansi>`,
		`<ansi fg="magenta">%s</ansi>`,
		`<ansi fg="white-bold">%s</ansi>`,
	}

	tFormat := string(configs.GetTextFormatsConfig().TimeShort)

	for itm := range user.EventLog.Items {

		if rest != `` && rest != itm.Category {
			continue
		}

		rows = append(rows, []string{
			itm.Category,
			//fmt.Sprintf(`%d`, itm.WhenRound),
			itm.WhenTime.Format(tFormat),
			itm.What,
		})

	}

	searchResultsTable := templates.GetTable(`History`, headers, rows, formatting)
	tplTxt, _ := templates.Process("tables/generic", searchResultsTable)
	user.SendText(tplTxt)

	return true, nil
}
