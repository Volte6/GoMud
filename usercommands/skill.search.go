package usercommands

import (
	"fmt"
	"math"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Search(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	skillLevel := user.Character.GetSkillLevel(skills.Search)

	if skillLevel == 0 {
		response.SendUserMessage(userId, "You don't know how to search.", true)
		response.Handled = true
		return response, fmt.Errorf("you don't know how to search")
	}

	if !user.Character.TryCooldown(skills.Search.String(), 2) {
		response.SendUserMessage(userId,
			fmt.Sprintf("You need to wait %d more rounds to use that skill again.", user.Character.GetCooldown(skills.Search.String())),
			true)
		response.Handled = true
		return response, fmt.Errorf("you're doing that too often")
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	// 10% + 1% for every 2 smarts
	searchOddsIn100 := 5 + int(math.Ceil(float64(user.Character.Stats.Smarts.Value)/4)) + int(math.Ceil(float64(user.Character.Stats.Perception.Value)/2))
	if skillLevel > 3 {
		searchOddsIn100 *= 2
	}

	response.SendUserMessage(userId, "You snoop around for a bit...\n", true)
	response.SendRoomMessage(user.Character.RoomId,
		fmt.Sprintf(`<ansi fg="username">%s</ansi> is snooping around.`, user.Character.Name),
		true)

	// Check room exists
	for exit, exitInfo := range room.Exits {
		if exitInfo.Secret {

			roll := util.Rand(100)

			util.LogRoll(`Secret Exit`, roll, searchOddsIn100)

			if roll < searchOddsIn100 {
				response.SendUserMessage(userId, fmt.Sprintf(`You found a secret exit: <ansi fg="secret-exit">%s</ansi>`, exit), true)
			}
		}
	}

	if skillLevel > 2 {
		// Find stashed items
		stashedItems := []string{}
		for _, item := range room.Stash {
			if !item.IsValid() {
				room.RemoveItem(item, true)
			}
			name := item.DisplayName() + ` <ansi fg="black-bold">(stashed)</ansi>`
			stashedItems = append(stashedItems, name)
		}

		hiddenPlayers := []string{}

		for _, pId := range room.GetPlayers() {
			if pId == userId {
				continue
			}
			if p := users.GetByUserId(pId); p != nil {

				roll := util.Rand(100)

				util.LogRoll(`Hidden Player`, roll, searchOddsIn100)

				if roll < searchOddsIn100 {
					if p.Character.HasBuffFlag(buffs.Hidden) {
						hiddenPlayers = append(hiddenPlayers, p.Character.Name+` <ansi fg="black-bold">(hiding)</ansi>`)
					}
				}
			}
		}

		if len(hiddenPlayers) > 0 {

			details := room.GetRoomDetails(user)
			details.VisiblePlayers = []characters.FormattedName{}

			for _, name := range hiddenPlayers {
				details.VisiblePlayers = append(details.VisiblePlayers,
					characters.FormattedName{
						Name:   name,
						Type:   `username`,
						Suffix: `hidden`,
					},
				)
			}

			whoTxt, _ := templates.Process("descriptions/who", details)
			response.SendUserMessage(userId, whoTxt, false)

		}

		hiddenMobs := []string{}

		for _, mId := range room.GetMobs() {
			if m := users.GetByUserId(mId); m != nil {

				roll := util.Rand(100)

				util.LogRoll(`Hidden Mob`, roll, searchOddsIn100)

				if roll < searchOddsIn100 {
					if m.Character.HasBuffFlag(buffs.Hidden) {
						hiddenMobs = append(hiddenPlayers, m.Character.Name+` <ansi fg="black-bold">(hiding)</ansi>`)
					}
				}
			}
		}

		if len(hiddenMobs) > 0 {

			details := room.GetRoomDetails(user)
			details.VisiblePlayers = []characters.FormattedName{}

			for _, name := range hiddenMobs {
				details.VisibleMobs = append(details.VisiblePlayers,
					characters.FormattedName{
						Name:   name,
						Type:   `mob`,
						Suffix: `hidden`,
					},
				)
			}

			whoTxt, _ := templates.Process("descriptions/who", details)
			response.SendUserMessage(userId, whoTxt, false)

		}

		//stashedItems := map[string][]string{}
		//stashedItems["Stashed here:"] = room.Stash
		textOut, _ := templates.Process("descriptions/ontheground", stashedItems)
		response.SendUserMessage(userId, textOut, false)
	}

	if skillLevel >= 3 {
		// Find props

	}

	response.Handled = true
	return response, nil
}
