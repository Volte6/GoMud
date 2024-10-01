package usercommands

import (
	"fmt"

	"github.com/volte6/mud/events"
	"github.com/volte6/mud/term"
	"github.com/volte6/mud/users"
)

// Global chat room
func Broadcast(rest string, user *users.UserRecord) (bool, error) {

	events.AddToQueue(events.Broadcast{
		Text: fmt.Sprintf(`<ansi fg="black-bold">(broadcast)</ansi> <ansi fg="username">%s</ansi>: <ansi fg="yellow">%s</ansi>%s`, user.Character.Name, rest, term.CRLFStr),
	})

	return true, nil
}
