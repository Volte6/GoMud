package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Inbox(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	tFormat := string(configs.GetConfig().TimeFormat)

	if rest == `clear` {
		user.Inbox.Empty()
	}

	if rest == `check` {
		user.SendText(fmt.Sprintf(`<ansi fg="159">You have <ansi fg="alert-4">%d</ansi> messages. Type <ansi fg="command">inbox</ansi> to view your messages.</ansi>`, len(user.Inbox)))
		return true, nil
	}

	user.SendText(fmt.Sprintf(`<ansi fg="159">You have <ansi fg="alert-4">%d</ansi> messages.</ansi>`, len(user.Inbox)))

	if len(user.Inbox) == 0 {
		return true, nil
	}

	user.SendText(`<ansi fg="105">` + strings.Repeat(`_`, 80) + `</ansi>`)

	for _, msg := range user.Inbox {
		user.SendText(``)
		user.SendText(fmt.Sprintf(`<ansi fg="51">Sent:</ansi>    <ansi fg="magenta">%s</ansi>`, msg.DateSent.Format(tFormat)))
		user.SendText(fmt.Sprintf(`<ansi fg="51">From:</ansi>    <ansi fg="username">%s</ansi>`, msg.FromName))
		user.SendText(`<ansi fg="51">Message:</ansi> <ansi fg="229">` + util.SplitStringNL(msg.Message, 71, `         `) + `</ansi>`)
		user.SendText(`<ansi fg="105">` + strings.Repeat(`_`, 80) + `</ansi>`)
	}

	user.SendText(``)
	user.SendText(`<ansi fg="159">Type <ansi fg="command">inbox clear</ansi> to clear all messages in your inbox.</ansi>`)
	user.SendText(``)

	return true, nil
}
