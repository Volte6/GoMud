package usercommands

import (
	"fmt"
	"strconv"

	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Status(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	possibleStatuses := []string{`strength`, `speed`, `smarts`, `vitality`, `mysticism`, `perception`}

	if rest != `` {

		args := util.SplitButRespectQuotes(rest)

		if len(args) < 2 || args[0] != `train` {
			user.SendText("stat WHAT???")
			response.Handled = true
			return response, nil
		}

		match, partial := util.FindMatchIn(args[1], possibleStatuses...)
		if len(match) == 0 {
			match = partial
		}

		trainQty := 0
		if len(args) > 2 {
			trainQty, _ = strconv.Atoi(args[2])
		}
		if trainQty < 1 {
			trainQty = 1
		}

		if user.Character.StatPoints < trainQty {
			user.SendText("You don't have enough stat points to do that.")
			response.Handled = true
			return response, nil
		}

		if len(match) == 0 {
			user.SendText("It's not clear which stat you want to improve.")
			response.Handled = true
			return response, nil
		}

		before := 0
		after := 0

		switch match {
		case `strength`:
			before = user.Character.Stats.Strength.Training
			user.Character.Stats.Strength.Training += trainQty
		case `speed`:
			before = user.Character.Stats.Speed.Training
			user.Character.Stats.Speed.Training += trainQty
		case `smarts`:
			before = user.Character.Stats.Smarts.Training
			user.Character.Stats.Smarts.Training += trainQty
		case `vitality`:
			before = user.Character.Stats.Vitality.Training
			user.Character.Stats.Vitality.Training += trainQty
		case `mysticism`:
			before = user.Character.Stats.Mysticism.Training
			user.Character.Stats.Mysticism.Training += trainQty
		case `perception`:
			before = user.Character.Stats.Perception.Training
			user.Character.Stats.Perception.Training += trainQty
		}

		after = before + trainQty
		user.Character.StatPoints -= trainQty

		user.Character.Validate()

		user.SendText(
			fmt.Sprintf(`Your base <ansi fg="yellow">%s</ansi> improves from <ansi fg="cyan">%d</ansi> to <ansi fg="cyan-bold">%d</ansi>!`, match, before, after))
		response.Handled = true
		return response, nil
	}

	tplTxt, _ := templates.Process("character/status", user)
	user.SendText(tplTxt)

	response.NextCommand = "inventory"
	response.Handled = true
	return response, nil
}
