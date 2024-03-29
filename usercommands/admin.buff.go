package usercommands

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/util"

	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
)

func Buff(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	// args should look like one of the following:
	// target buffId - put buff on target if in the room
	// buffId - put buff on self
	// search searchTerm - search for buff by name, display results
	args := util.SplitButRespectQuotes(rest)

	if len(args) > 0 {

		if (len(args) >= 2 && args[0] == "search") || (len(args) == 1 && args[0] == "list") {

			var foundBuffIds []int

			if args[0] == "list" {
				foundBuffIds = buffs.GetAllBuffIds()
			} else {
				foundBuffIds = buffs.SearchBuffs(args[1])
			}

			sort.Ints(foundBuffIds)

			headers := []string{"Id", "Description", "Flags"}
			rows := [][]string{}

			if len(foundBuffIds) == 0 {
				rows = append(rows, []string{"No Matches", "No Matches", "No Matches"})
			} else {
				for _, buffId := range foundBuffIds {
					if buffSpec := buffs.GetBuffSpec(buffId); buffSpec != nil {
						flags := []string{}
						for _, flag := range buffSpec.Flags {
							flags = append(flags, string(flag))
						}
						rows = append(rows, []string{strconv.Itoa(buffSpec.BuffId), buffSpec.Name, strings.Join(flags, ", ")})
						rows = append(rows, []string{``, `-` + buffSpec.Description, ``})
					}
				}
			}

			searchResultsTable := templates.GetTable("Search Results", headers, rows)
			tplTxt, _ := templates.Process("tables/generic", searchResultsTable)
			response.SendUserMessage(userId, tplTxt, false)
		} else {

			targetUserId := 0
			targetMobInstanceId := 0
			buffId := 0

			if len(args) >= 2 {

				room := rooms.LoadRoom(user.Character.RoomId)
				if room == nil {
					return response, fmt.Errorf(`room %d not found`, user.Character.RoomId)
				}

				targetUserId, targetMobInstanceId = room.FindByName(args[0])

				buffId, _ = strconv.Atoi(args[1])
				if buffId == 0 {
					// Grab the first match
					foundBuffIds := buffs.SearchBuffs(args[1])
					if len(foundBuffIds) > 0 {
						buffId = foundBuffIds[0]
					}
				}

			} else if len(args) == 1 {
				targetUserId = userId
				buffId, _ = strconv.Atoi(args[0])
				if buffId == 0 {
					// Grab the first match
					foundBuffIds := buffs.SearchBuffs(args[0])
					if len(foundBuffIds) > 0 {
						buffId = foundBuffIds[0]
					}
				}
			}

			if buffId == 0 {
				response.SendUserMessage(userId, "buffId must be an integer > 0.", true)
				response.Handled = true
				return response, nil

			}

			if targetUserId > 0 {
				// get the user
				if targetUser := users.GetByUserId(targetUserId); targetUser != nil {
					// Get the buff
					if buffSpec := buffs.GetBuffSpec(buffId); buffSpec != nil {

						// Apply the buff
						cmdQueue.QueueBuff(targetUserId, 0, buffId)
						response.SendUserMessage(userId, fmt.Sprintf("Buff %d (%s) applied to %s.", buffId, buffSpec.Name, targetUser.Character.Name), true)

					} else {
						response.SendUserMessage(userId, fmt.Sprintf("Buff Id %d not found.", buffId), true)
					}

					response.Handled = true
					return response, nil
				}
			}

			if targetMobInstanceId > 0 {
				// get the user
				if targetMob := mobs.GetInstance(targetMobInstanceId); targetMob != nil {
					// Get the buff
					if buffSpec := buffs.GetBuffSpec(buffId); buffSpec != nil {

						// Apply the buff
						cmdQueue.QueueBuff(0, targetMobInstanceId, buffSpec.BuffId)

						response.SendUserMessage(userId, fmt.Sprintf("Buff %d (%s) applied to %s.", buffSpec.BuffId, buffSpec.Name, targetMob.Character.Name), true)

					} else {
						response.SendUserMessage(userId, fmt.Sprintf("Buff Id %d not found.", buffId), true)
					}

					response.Handled = true
					return response, nil
				}
			}

		}
	}

	response.SendUserMessage(userId, "target not found.", true)

	// send some sort of help info?
	infoOutput, _ := templates.Process("admincommands/help/command.buff", nil)
	response.SendUserMessage(userId, infoOutput, false)

	response.Handled = true
	return response, nil
}
