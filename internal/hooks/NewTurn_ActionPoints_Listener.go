package hooks

import (
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/users"
)

//
// Update movement points for each player
// TODO: Optimize this to avoid re-loops through users
//

func ActionPoints_Listener(e events.Event) bool {
	//evt := e.(events.NewTurn)

	for _, user := range users.GetAllActiveUsers() {
		user.Character.ActionPoints += 1
		if user.Character.ActionPoints > user.Character.ActionPointsMax.Value {
			user.Character.ActionPoints = user.Character.ActionPointsMax.Value
		}
	}

	return true
}
