package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
)

func Inbox(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	if rest == `clear` {
		user.Inbox.Empty()
	}

	if rest == `check` {
		user.SendText(fmt.Sprintf(`<ansi fg="159">You have <ansi fg="alert-4">%d</ansi> unread messages and <ansi fg="alert-4">%d</ansi> old messages. Type <ansi fg="command">inbox</ansi> to view your messages.</ansi>`, user.Inbox.CountUnread(), user.Inbox.CountRead()))
		return true, nil
	}

	user.SendText(fmt.Sprintf(`<ansi fg="159">You have <ansi fg="alert-4">%d</ansi> unread messages and <ansi fg="alert-4">%d</ansi> old messages.</ansi>`, user.Inbox.CountUnread(), user.Inbox.CountRead()))

	if len(user.Inbox) == 0 {
		return true, nil
	}

	border := `<ansi fg="mail-border">` + strings.Repeat(`_`, 80) + `</ansi>`
	user.SendText(border)

	for idx, msg := range user.Inbox {

		if rest == `old` {
			if !msg.Read {
				continue
			}
		} else if msg.Read {
			continue
		}

		tplTxt, _ := templates.Process("mail/message", msg)
		user.SendText(tplTxt)

		user.SendText(border)

		if !msg.Read {
			if msg.Gold > 0 {
				user.Character.Bank += msg.Gold
			}
			if msg.Item != nil {
				user.Character.StoreItem(*msg.Item)
			}
		}

		user.Inbox[idx].Read = true
	}

	user.SendText(``)
	user.SendText(`<ansi fg="159">Type <ansi fg="command">inbox old</ansi> to read old messages.</ansi>`)
	user.SendText(`<ansi fg="159">Type <ansi fg="command">inbox clear</ansi> to clear all messages in your inbox.</ansi>`)
	user.SendText(``)

	return true, nil
}
