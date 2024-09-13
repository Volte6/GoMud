package usercommands

import (
	"errors"
	"fmt"

	"github.com/volte6/mud/items"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

/*
Peep Skill
Level 1 - Reveals the type and value of items.
Level 2 - Reveals weapon damage or uses an item has left.
Level 3 - Reveals any stat modifiers an item has.
Level 4 - Reveals special magical properties like elemental effects.
*/
func Inspect(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	if user.Character.GetSkillLevel(skills.Inspect) == 0 {
		response.SendUserMessage(userId, "You don't know how to inspect.")
		response.Handled = true
		return response, fmt.Errorf("you don't know how to inspect")
	}

	if len(rest) == 0 {
		response.SendUserMessage(userId, "Type `help inspect` for more information on the inspect skill.")
		response.Handled = true
		return response, nil
	}

	skillLevel := user.Character.GetSkillLevel(skills.Inspect)

	// Check whether the user has an item in their inventory that matches
	matchItem, found := user.Character.FindInBackpack(rest)

	if !found {
		response.SendUserMessage(userId, fmt.Sprintf("You don't have a %s to inspect. Is it still worn, perhaps?", rest))
	} else {

		if !user.Character.TryCooldown(skills.Inspect.String(), 3) {
			response.SendUserMessage(userId,
				fmt.Sprintf("You need to wait %d more rounds to use that skill again.", user.Character.GetCooldown(skills.Inspect.String())),
			)
			response.Handled = true
			return response, errors.New(`you're doing that too often`)
		}

		response.SendUserMessage(userId,
			fmt.Sprintf(`You inspect the <ansi fg="item">%s</ansi>.`, matchItem.DisplayName()),
		)
		response.SendRoomMessage(user.Character.RoomId,
			fmt.Sprintf(`<ansi fg="username">%s</ansi> inspects their <ansi fg="item">%s</ansi>...`, user.Character.Name, matchItem.DisplayName()),
		)

		type inspectDetails struct {
			InspectLevel int
			Item         *items.Item
			ItemSpec     *items.ItemSpec
		}

		iSpec := matchItem.GetSpec()

		details := inspectDetails{
			InspectLevel: skillLevel,
			Item:         &matchItem,
			ItemSpec:     &iSpec,
		}

		inspectTxt, _ := templates.Process("descriptions/inspect", details)
		response.SendUserMessage(userId, inspectTxt)

	}

	response.Handled = true
	return response, nil
}
