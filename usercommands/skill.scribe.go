package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/items"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Scribe(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	skillLevel := user.Character.GetSkillLevel(skills.Script)

	if skillLevel == 0 {
		response.SendUserMessage(userId, "You don't know how to scribe.", true)
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
	// map - draw a map of the room and write it to an item
	args := util.SplitButRespectQuotes(rest)

	if len(args) == 0 {
		response.SendUserMessage(userId, "Type `help scribe` for more information on the scribe skill.", true)
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

		response.SendUserMessage(userId, "You write a note, and tuck it away safely.", true)

	} else if scribeType == "sign" {

		if skillLevel < 2 {

			response.SendUserMessage(userId, "You don't know how to create signs yet.", true)

		} else if !user.Character.TryCooldown(skills.Scribe.String(), 10) {

			// There's a cooldown on this skill
			response.SendUserMessage(userId,
				fmt.Sprintf("You need to wait %d more rounds to use that skill again.", user.Character.GetCooldown(skills.Scribe.String())),
				true)
			response.Handled = true
			return response, fmt.Errorf("you're doing that too often")

		} else {
			// Write a sign in the room
			if len(rest) > 50 {
				response.SendUserMessage(userId, "That won't fit! Keep it under 50 letters.", true)
			} else {
				if replaced := room.AddSign(rest, 0, 7); replaced {
					response.SendUserMessage(userId, "You knock down the old sign and replace it with a new one.", true)
					response.SendRoomMessage(user.Character.RoomId,
						fmt.Sprintf(`<ansi fg="username">%s</ansi> knocks down the old sign and replaces it with a new one.`, user.Character.Name),
						true)
				} else {
					response.SendUserMessage(userId, "You find some junk wood and scrawl a message onto it.", true)
					response.SendRoomMessage(user.Character.RoomId,
						fmt.Sprintf(`<ansi fg="username">%s</ansi> finds some junk wood and scrawls a message onto it.`, user.Character.Name),
						true)
				}
			}
		}
	} else if scribeType == "rune" {

		if skillLevel < 3 {

			response.SendUserMessage(userId, "You don't know how to create runes yet.", true)

		} else if !user.Character.TryCooldown(skills.Scribe.String(), 2) {

			// There's a cooldown on this skill
			response.SendUserMessage(userId,
				fmt.Sprintf("You need to wait %d more rounds to use that skill again.", user.Character.GetCooldown(skills.Scribe.String())),
				true)
			response.Handled = true
			return response, fmt.Errorf("you're doing that too often")

		} else {

			// Write a rune in the room
			if len(rest) > 50 {
				response.SendUserMessage(userId, "That won't fit! Keep it under 50 letters.", true)
			} else {
				if replaced := room.AddSign(rest, userId, 7); replaced {
					response.SendUserMessage(userId, "You scratch out the old rune and replace it with a new one.", true)
				} else {
					response.SendUserMessage(userId, "You scratch a rune into the floor.", true)
				}
			}

		}

	} else if scribeType == "map" {

		if skillLevel < 4 {

			response.SendUserMessage(userId, "You don't know how to scribe maps yet.", true)

		} else if !user.Character.TryCooldown(skills.Scribe.String(), 30) {

			// There's a cooldown on this skill
			response.SendUserMessage(userId,
				fmt.Sprintf("You need to wait %d more rounds to use that skill again.", user.Character.GetCooldown(skills.Scribe.String())),
				true)
			response.Handled = true
			return response, fmt.Errorf("you're doing that too often")

		} else {
			// Draw a map of the room and write it to an item
			resp, err := Map("", userId, cmdQueue)
			if err != nil {
				response.SendUserMessage(userId, err.Error(), true)
				return response, err
			}

			mapContents := resp.GetUserMessagesAsString(userId)
			mapContents = strings.Replace(mapContents, "@", "X", -1)
			mapContents = strings.Replace(mapContents, "You", "Here", -1)

			mapItem := items.New(2)
			mapItem.SetBlob(templates.AnsiParse(mapContents))
			user.Character.StoreItem(mapItem)

			response.SendUserMessage(userId, "You draw a map of the area, as much as you can remember it.", true)
		}
	}

	response.Handled = true
	return response, nil
}
