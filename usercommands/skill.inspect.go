package usercommands

import (
	"errors"
	"fmt"

	"github.com/volte6/mud/items"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
)

/*
Peep Skill
Level 1 - Reveals the type and value of items.
Level 2 - Reveals weapon damage or uses an item has left.
Level 3 - Reveals any stat modifiers an item has.
Level 4 - Reveals special magical properties like elemental effects.
*/
func Inspect(rest string, userId int) (bool, string, error) {

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return false, ``, fmt.Errorf("user %d not found", userId)
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return false, ``, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	if user.Character.GetSkillLevel(skills.Inspect) == 0 {
		user.SendText("You don't know how to inspect.")
		return true, ``, fmt.Errorf("you don't know how to inspect")
	}

	if len(rest) == 0 {
		user.SendText("Type `help inspect` for more information on the inspect skill.")
		return true, ``, nil
	}

	skillLevel := user.Character.GetSkillLevel(skills.Inspect)

	// Check whether the user has an item in their inventory that matches
	matchItem, found := user.Character.FindInBackpack(rest)

	if !found {
		user.SendText(fmt.Sprintf("You don't have a %s to inspect. Is it still worn, perhaps?", rest))
	} else {

		if !user.Character.TryCooldown(skills.Inspect.String(), 3) {
			user.SendText(
				fmt.Sprintf("You need to wait %d more rounds to use that skill again.", user.Character.GetCooldown(skills.Inspect.String())),
			)
			return true, ``, errors.New(`you're doing that too often`)
		}

		user.SendText(
			fmt.Sprintf(`You inspect the <ansi fg="item">%s</ansi>.`, matchItem.DisplayName()),
		)
		room.SendText(
			fmt.Sprintf(`<ansi fg="username">%s</ansi> inspects their <ansi fg="item">%s</ansi>...`, user.Character.Name, matchItem.DisplayName()),
			userId,
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
		user.SendText(inspectTxt)

	}

	return true, ``, nil
}
