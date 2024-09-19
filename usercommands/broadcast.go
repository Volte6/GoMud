package usercommands

import (
	"fmt"

	"github.com/volte6/mud/events"
	"github.com/volte6/mud/term"
	"github.com/volte6/mud/users"
)

// Global chat room
func Broadcast(rest string, userId int) (bool, error) {

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return false, fmt.Errorf("user %d not found", userId)
	}

	events.AddToQueue(events.Broadcast{
		Text: fmt.Sprintf(`<ansi fg="black" bold="true'>(broadcast)</ansi> <ansi fg="username">%s</ansi>: <ansi fg="yellow">%s</ansi>%s`, user.Character.Name, rest, term.CRLFStr),
	})

	return true, nil
}
