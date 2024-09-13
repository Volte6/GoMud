package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/items"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

/*
Scribe Skill
Level 1 - Scribe to a scrap of paper
Level 2 - Scribe to a sign
Level 3 - Scribe a hidden rune
Level 4 - TODO
*/
func Scribe(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	skillLevel := user.Character.GetSkillLevel(skills.Scribe)

	if skillLevel == 0 {
		user.SendText("You don't know how to scribe.")
		response.Handled = true
		return response, fmt.Errorf("you don't know how to scribe")
	}

	// Load current room details

	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	// args should look like one of the following:
	// note a bunch of text that follows - write a note and create an item of it
	// sign a bunch of text that follows - scratch a message on a sign in the room
	// rune some secret text that only the user should see - scratch a private rune message
	//
	args := util.SplitButRespectQuotes(rest)

	if len(args) == 0 {
		user.SendText("Type `help scribe` for more information on the scribe skill.")
		response.Handled = true
		return response, nil
	}

	scribeType := args[0]
	rest = strings.Join(args[1:], " ")

	if scribeType == "note" {
		// Create a note item
		noteItem := items.New(1)
		noteItem.SetBlob(rest)
		user.Character.StoreItem(noteItem)

		user.SendText("You write a note, and tuck it away safely.")

	} else if scribeType == "sign" {

		if skillLevel < 2 {

			user.SendText("You don't know how to create signs yet.")

		} else if !user.Character.TryCooldown(skills.Scribe.String(), 10) {

			// There's a cooldown on this skill
			user.SendText(
				fmt.Sprintf("You need to wait %d more rounds to use that skill again.", user.Character.GetCooldown(skills.Scribe.String())),
			)
			response.Handled = true
			return response, fmt.Errorf("you're doing that too often")

		} else {
			// Write a sign in the room
			if len(rest) > 50 {
				user.SendText("That won't fit! Keep it under 50 letters.")
			} else {
				if replaced := room.AddSign(rest, 0, 7); replaced {
					user.SendText("You knock down the old sign and replace it with a new one.")
					room.SendText(
						fmt.Sprintf(`<ansi fg="username">%s</ansi> knocks down the old sign and replaces it with a new one.`, user.Character.Name),
						userId,
					)
				} else {
					user.SendText("You find some junk wood and scrawl a message onto it.")
					room.SendText(
						fmt.Sprintf(`<ansi fg="username">%s</ansi> finds some junk wood and scrawls a message onto it.`, user.Character.Name),
						userId,
					)
				}
			}
		}
	} else if scribeType == "rune" {

		if skillLevel < 3 {

			user.SendText("You don't know how to create runes yet.")

		} else if !user.Character.TryCooldown(skills.Scribe.String(), 2) {

			// There's a cooldown on this skill
			user.SendText(
				fmt.Sprintf("You need to wait %d more rounds to use that skill again.", user.Character.GetCooldown(skills.Scribe.String())),
			)
			response.Handled = true
			return response, fmt.Errorf("you're doing that too often")

		} else {

			// Write a rune in the room
			if len(rest) > 50 {
				user.SendText("That won't fit! Keep it under 50 letters.")
			} else {
				if replaced := room.AddSign(rest, userId, 7); replaced {
					user.SendText("You scratch out the old rune and replace it with a new one.")
				} else {
					user.SendText("You scratch a rune into the floor.")
				}
			}

		}

	}

	response.Handled = true
	return response, nil
}
