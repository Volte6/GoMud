package usercommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func RoomProps(cmd string, rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

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

	if feedbackStr, ok := room.TryLookProp(cmd, rest); ok {
		if feedbackStr != "" {
			response.SendUserMessage(userId, fmt.Sprintf(`<ansi fg="description">%s</ansi>`, util.SplitStringNL(feedbackStr, 80)), true)
			response.Handled = true
		}
		if !isSneaking {
			response.SendRoomMessage(room.RoomId, fmt.Sprintf("%s looks at the %s.", user.Character.Name, rest), true)
		}
	}

	trigger, oncooldown, rejectionMessage := room.TryTrigger(cmd, rest, userId)
	if len(rejectionMessage) > 0 {
		response.Handled = true
		response.SendUserMessage(userId, rejectionMessage, true)
	}

	if trigger == nil {
		return response, nil
	}

	if oncooldown {
		response.Handled = true
		// need to send a message to the user
		response.SendUserMessage(userId, "You need to wait a while before that will work again.", true)
		return response, nil
	}

	user.Character.CancelBuffsWithFlag(buffs.Hidden)

	// Send descriptions to the player and/or entire room
	if len(trigger.DescriptionPlayer) > 0 {
		response.SendUserMessage(userId, util.SplitStringNL(trigger.DescriptionPlayer, 80), true)
	}
	if len(trigger.DescriptionRoom) > 0 {
		response.SendRoomMessage(room.RoomId, util.SplitStringNL(fmt.Sprintf(trigger.DescriptionRoom, user.Character.Name), 80), true)
	}

	affectedPlayerIds := []int{}
	if trigger.Affected == rooms.AffectsPlayer {
		affectedPlayerIds = append(affectedPlayerIds, userId)
	} else if trigger.Affected == rooms.AffectsRoom {
		affectedPlayerIds = room.GetPlayers()
	}

	for _, affectedPlayerId := range affectedPlayerIds {

		triggerUser := users.GetByUserId(affectedPlayerId)
		if triggerUser == nil {
			continue
		}

		// Execute the trigger
		if trigger.RoomId > 0 {
			// The trigger does something, so mark it as handled
			// If we don't do this, it allows the command to pass through to the userCommands map
			response.Handled = true

			destRoom := rooms.LoadRoom(trigger.RoomId)

			// move the player
			rooms.MoveToRoom(triggerUser.UserId, trigger.RoomId)

			for _, mobInstId := range room.GetMobs(rooms.FindCharmed) {
				if m := mobs.GetInstance(mobInstId); m != nil {
					if m.Character.IsCharmed(triggerUser.UserId) {
						response.SendUserMessage(triggerUser.UserId, fmt.Sprintf("%s slips in behind you.", m.Character.Name), true)
						response.SendRoomMessage(room.RoomId, fmt.Sprintf(`<ansi fg="mobname">%s</ansi> slips in behind <ansi fg="username">%s</ansi>.`, m.Character.Name, user.Character.Name), true)
						room.RemoveMob(mobInstId)
						destRoom.AddMob(mobInstId)
					}
				}
			}
		}

		if trigger.BuffId != 0 {
			// The trigger does something, so mark it as handled
			// If we don't do this, it allows the command to pass through to the userCommands map
			response.Handled = true

			// Give the player a buff
			cmdQueue.QueueBuff(triggerUser.UserId, 0, trigger.BuffId)
		}

		// Give them a quest token
		if len(trigger.QuestToken) > 0 {
			// The trigger does something, so mark it as handled
			// If we don't do this, it allows the command to pass through to the userCommands map
			response.Handled = true

			// Give them a token
			cmdQueue.QueueQuest(triggerUser.UserId, trigger.QuestToken)
		}

		if trigger.MapInfo != `` {
			// map arguments are:
			// roomid:[1]/size:[wide/normal]/secrets:false/height:[18]/name:[Map of Frostfang]
			mapTxt := rooms.GetMapForDataString(trigger.MapInfo)
			response.SendUserMessage(triggerUser.UserId, mapTxt, false)
			response.Handled = true
		}

		if trigger.ItemId > 0 {
			// The trigger does something, so mark it as handled
			// If we don't do this, it allows the command to pass through to the userCommands map
			response.Handled = true

			newItem := items.New(trigger.ItemId)
			triggerUser.Character.StoreItem(newItem)

			iSpec := newItem.GetSpec()
			if iSpec.QuestToken != `` {
				cmdQueue.QueueQuest(triggerUser.UserId, iSpec.QuestToken)
			}
		}

		if len(trigger.SkillInfo) > 0 {
			response.Handled = true

			details := strings.Split(trigger.SkillInfo, `:`)
			if len(details) > 1 {
				skillName := strings.ToLower(details[0])
				skillLevel, _ := strconv.Atoi(details[1])
				currentLevel := triggerUser.Character.GetSkillLevel(skills.SkillTag(skillName))

				if currentLevel < skillLevel {
					newLevel := triggerUser.Character.TrainSkill(skillName, skillLevel)

					skillData := struct {
						SkillName  string
						SkillLevel int
					}{
						SkillName:  skillName,
						SkillLevel: newLevel,
					}
					skillUpTxt, _ := templates.Process("character/skillup", skillData)

					response.SendUserMessage(triggerUser.UserId, skillUpTxt, true)
				}

			}
		}

	}

	return response, nil
}
