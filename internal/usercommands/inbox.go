package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/language"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
)

func Inbox(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if rest == `clear` {
		user.Inbox.Empty()
	}

	if rest == `check` {
		user.SendText(fmt.Sprintf(language.T(`Inbox.UnreadMessageWithCheck`), user.Inbox.CountUnread(), user.Inbox.CountRead()))
		return true, nil
	}

	user.SendText(fmt.Sprintf(language.T(`Inbox.UnreadMessage`), user.Inbox.CountUnread(), user.Inbox.CountRead()))

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

		tplTxt, _ := templates.Process("mail/message", msg, user.UserId)
		user.SendText(tplTxt)

		user.SendText(border)

		if !msg.Read {
			if msg.Gold > 0 {
				user.Character.Bank += msg.Gold

				events.AddToQueue(events.EquipmentChange{
					UserId:     user.UserId,
					BankChange: msg.Gold,
				})

			}
			if msg.Item != nil {
				user.Character.StoreItem(*msg.Item)
			}
		}

		user.Inbox[idx].Read = true
	}

	user.SendText(``)
	user.SendText(language.T(`Inbox.ReadOldMessages`))
	user.SendText(language.T(`Inbox.ClearMessages`))
	user.SendText(``)

	return true, nil
}
