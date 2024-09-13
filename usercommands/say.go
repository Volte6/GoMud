package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Say(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	isSneaking := user.Character.HasBuffFlag(buffs.Hidden)
	isDrunk := user.Character.HasBuffFlag(buffs.Drunk)

	if isDrunk {
		// modify the text to look like it's the speech of a drunk person
		rest = drunkify(rest)
	}
	if isSneaking {
		room.SendText(fmt.Sprintf(`someone says, "<ansi fg="yellow">%s</ansi>"`, rest), userId)
	} else {
		room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> says, "<ansi fg="yellow">%s</ansi>"`, user.Character.Name, rest), userId)
	}

	user.SendText(fmt.Sprintf(`You say, "<ansi fg="yellow">%s</ansi>"`, rest))

	response.Handled = true
	return response, nil
}

func drunkify(sentence string) string {

	var drunkSentence strings.Builder
	isStartOfWord := true
	sentenceLength := len(sentence)
	insertedHiccup := false

	for i, char := range sentence {
		// Randomly decide whether to modify the character
		if util.Rand(10) < 3 || (!insertedHiccup || i == sentenceLength-1) {
			switch char {
			case 's':
				if isStartOfWord {
					drunkSentence.WriteString("sss")
				} else {
					drunkSentence.WriteString("sh")
				}
			case 'S':
				drunkSentence.WriteString("Sh")
			default:
				drunkSentence.WriteRune(char)
			}

			// Insert a hiccup in the middle of the sentence
			if !insertedHiccup && i >= sentenceLength/2 {
				drunkSentence.WriteString(" *hiccup* ")
				insertedHiccup = true
			}
		} else {
			drunkSentence.WriteRune(char)
		}

		// Update isStartOfWord based on spaces and punctuation
		if char == ' ' || char == '.' || char == '!' || char == '?' || char == ',' {
			isStartOfWord = true
		} else {
			isStartOfWord = false
		}
	}

	return drunkSentence.String()
}
