package hooks

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/quests"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/skills"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/users"
)

//
// Handles quest progress
//

func HandleQuestUpdate(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.Quest)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "Quest", "Actual Type", e.Type())
		return events.Cancel
	}

	//mudlog.Debug(`Event`, `type`, evt.Type(), `UserId`, evt.UserId, `QuestToken`, evt.QuestToken)

	// Give them a token
	remove := false
	if evt.QuestToken[0:1] == `-` {
		remove = true
		evt.QuestToken = evt.QuestToken[1:]
	}

	questInfo := quests.GetQuest(evt.QuestToken)
	if questInfo == nil {
		return events.Continue
	}

	questUser := users.GetByUserId(evt.UserId)
	if questUser == nil {
		return events.Continue
	}

	if remove {
		questUser.Character.ClearQuestToken(evt.QuestToken)
		return events.Continue
	}
	// This only succees if the user doesn't have the quest yet or the quest is a later step of one they've started
	if !questUser.Character.GiveQuestToken(evt.QuestToken) {
		return events.Continue
	}

	_, stepName := quests.TokenToParts(evt.QuestToken)
	if stepName == `start` {
		if !questInfo.Secret {

			questUser.EventLog.Add(`quest`, fmt.Sprintf(`Given a new quest: <ansi fg="questname">%s</ansi>`, questInfo.Name))

			questUpTxt, _ := templates.Process("character/questup", fmt.Sprintf(`You have been given a new quest: <ansi fg="questname">%s</ansi>!`, questInfo.Name), questUser.UserId)
			questUser.SendText(questUpTxt)
		}
	} else if stepName == `end` {

		if !questInfo.Secret {

			questUser.EventLog.Add(`quest`, fmt.Sprintf(`Completed a quest: <ansi fg="questname">%s</ansi>`, questInfo.Name))

			questUpTxt, _ := templates.Process("character/questup", fmt.Sprintf(`You have completed the quest: <ansi fg="questname">%s</ansi>!`, questInfo.Name), questUser.UserId)
			questUser.SendText(questUpTxt)
		}

		// Message to player?
		if len(questInfo.Rewards.PlayerMessage) > 0 {
			questUser.SendText(questInfo.Rewards.PlayerMessage)
		}
		// Message to room?
		if len(questInfo.Rewards.RoomMessage) > 0 {
			if room := rooms.LoadRoom(questUser.Character.RoomId); room != nil {
				room.SendText(questInfo.Rewards.RoomMessage, questUser.UserId)
			}
		}
		// New quest to start?
		if len(questInfo.Rewards.QuestId) > 0 {

			events.AddToQueue(events.Quest{
				UserId:     questUser.UserId,
				QuestToken: questInfo.Rewards.QuestId,
			})

		}
		// Gold reward?
		if questInfo.Rewards.Gold > 0 {
			questUser.SendText(fmt.Sprintf(`You receive <ansi fg="gold">%d gold</ansi>!`, questInfo.Rewards.Gold))
			questUser.Character.Gold += questInfo.Rewards.Gold

			events.AddToQueue(events.EquipmentChange{
				UserId:     questUser.UserId,
				GoldChange: questInfo.Rewards.Gold,
			})

		}
		// Item reward?
		if questInfo.Rewards.ItemId > 0 {
			newItm := items.New(questInfo.Rewards.ItemId)
			questUser.SendText(fmt.Sprintf(`You receive <ansi fg="itemname">%s</ansi>!`, newItm.NameSimple()))
			questUser.Character.StoreItem(newItm)

			iSpec := newItm.GetSpec()
			if iSpec.QuestToken != `` {

				events.AddToQueue(events.Quest{
					UserId:     questUser.UserId,
					QuestToken: iSpec.QuestToken,
				})

			}
		}
		// Buff reward?
		if questInfo.Rewards.BuffId > 0 {
			questUser.AddBuff(questInfo.Rewards.BuffId, `quest`)
		}
		// Experience reward?
		if questInfo.Rewards.Experience > 0 {
			questUser.GrantXP(questInfo.Rewards.Experience, `quest progress`)
		}
		// Skill reward?
		if questInfo.Rewards.SkillInfo != `` {
			details := strings.Split(questInfo.Rewards.SkillInfo, `:`)
			if len(details) > 1 {
				skillName := strings.ToLower(details[0])
				skillLevel, _ := strconv.Atoi(details[1])
				currentLevel := questUser.Character.GetSkillLevel(skills.SkillTag(skillName))

				if currentLevel < skillLevel {
					newLevel := questUser.Character.TrainSkill(skillName, skillLevel)

					skillData := struct {
						SkillName  string
						SkillLevel int
					}{
						SkillName:  skillName,
						SkillLevel: newLevel,
					}
					skillUpTxt, _ := templates.Process("character/skillup", skillData, questUser.UserId)
					questUser.SendText(skillUpTxt)
				}

			}
		}
		// Move them to another room/area?
		if questInfo.Rewards.RoomId > 0 {
			questUser.SendText(`You are suddenly moved to a new place!`)

			if room := rooms.LoadRoom(questUser.Character.RoomId); room != nil {
				room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> is suddenly moved to a new place!`, questUser.Character.Name), questUser.UserId)
			}

			rooms.MoveToRoom(questUser.UserId, questInfo.Rewards.RoomId)
		}
	} else {
		if !questInfo.Secret {

			questUser.EventLog.Add(`quest`, fmt.Sprintf(`Made progress on a quest: <ansi fg="questname">%s</ansi>`, questInfo.Name))

			questUpTxt, _ := templates.Process("character/questup", fmt.Sprintf(`You've made progress on the quest: <ansi fg="questname">%s</ansi>!`, questInfo.Name), questUser.UserId)
			questUser.SendText(questUpTxt)
		}
	}

	return events.Continue
}
