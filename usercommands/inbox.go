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

	border := `<ansi fg="105">` + strings.Repeat(`_`, 80) + `</ansi>`
	user.SendText(border)

	for idx, msg := range user.Inbox {

		dateColor := `135`
		sectionColor := `37`
		msgColor := `188`
		if !msg.Read {
			dateColor = `magenta`
			sectionColor = `51`
			msgColor = `229`
		}

		user.SendText(``)
		user.SendText(fmt.Sprintf(`<ansi fg="%s">Sent:</ansi>    <ansi fg="%s">%s</ansi>`, sectionColor, dateColor, msg.DateSent.Format(tFormat)))
		user.SendText(fmt.Sprintf(`<ansi fg="%s">From:</ansi>    <ansi fg="username">%s</ansi>`, sectionColor, msg.FromName))
		user.SendText(fmt.Sprintf(`<ansi fg="%s">Message:</ansi> <ansi fg="%s">`, sectionColor, msgColor) + util.SplitStringNL(msg.Message, 71, `         `) + `</ansi>`)

		if !msg.Read {

			if msg.Gold > 0 {
				user.SendText(fmt.Sprintf(`<ansi fg="alert-4">NOTE: </ansi> This message had <ansi fg="gold">%d gold</ansi> attached! It has been added to your bank balance.`, msg.Gold))
				user.Character.Bank += msg.Gold
			}

			if msg.Item.ItemId > 0 {
				user.SendText(fmt.Sprintf(`<ansi fg="alert-4">NOTE: </ansi> This message came with one <ansi fg="itemname">%s</ansi> attached! It has been added to your inventory.`, msg.Item.DisplayName()))
				user.Character.StoreItem(msg.Item)
			}
		}

		user.Inbox[idx].Read = true

		user.SendText(border)
	}

	user.SendText(``)
	user.SendText(`<ansi fg="159">Type <ansi fg="command">inbox clear</ansi> to clear all messages in your inbox.</ansi>`)
	user.SendText(``)

	return true, nil
}
