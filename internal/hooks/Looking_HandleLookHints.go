package hooks

import (
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
)

func HandleLookHints(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.Looking)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "Looking", "Actual Type", e.Type())
		return events.Cancel
	}

	if evt.Target == `` {

		user := users.GetByUserId(evt.UserId)
		if user == nil {
			return events.Cancel
		}

		if user.DidTip(`list`) {
			return events.Continue
		}

		room := rooms.LoadRoom(evt.RoomId)
		if room == nil {
			return events.Cancel
		}

		showListTip := false
		if len(room.GetMobs(rooms.FindMerchant)) > 0 {
			showListTip = true
		} else if len(room.GetPlayers(rooms.FindMerchant)) > 0 {
			showListTip = true
		}

		if showListTip {
			user.SendText(`<ansi fg="alert-5">TIP:</ansi> <ansi fg="tip-text">Type <ansi fg="command">list</ansi> to see what merchants have for sale.</ansi>`)
			user.SendText(``)
		}

	}
	return events.Continue
}
