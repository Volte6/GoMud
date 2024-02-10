package scripting

import (
	"strings"

	"github.com/dop251/goja"
	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
)

func setUserFunctions(vm *goja.Runtime) {
	vm.Set(`UserSetTempData`, UserSetTempData)
	vm.Set(`UserGetTempData`, UserGetTempData)
	vm.Set(`UserGetCharacterName`, UserGetCharacterName)
	vm.Set(`UserGetRoomId`, UserGetRoomId)
	vm.Set(`UserGetRoomId`, UserGetRoomId)
	vm.Set(`UserHasQuest`, UserHasQuest)
	vm.Set(`UserGiveQuest`, UserGiveQuest)
	vm.Set(`UserGiveBuff`, UserGiveBuff)
	vm.Set(`UserCommand`, UserCommand)
	vm.Set(`UserTrainSkill`, UserTrainSkill)
	vm.Set(`UserMoveRoom`, UserMoveRoom)
	vm.Set(`UserGiveItem`, UserGiveItem)
	vm.Set(`UserHasBuffFlag`, UserHasBuffFlag)
	vm.Set(`UserHasItemId`, UserHasItemId)
	vm.Set(`UserGetBackpackItems`, UserGetBackpackItems)
}

// ////////////////////////////////////////////////////////
//
// # These functions get exported to the scripting engine
//
// ////////////////////////////////////////////////////////
func UserSetTempData(userId int, key string, value any) {
	if user := users.GetByUserId(userId); user != nil {
		user.SetTempData(key, value)
	}
}

func UserGetTempData(userId int, key string) any {
	if user := users.GetByUserId(userId); user != nil {
		return user.GetTempData(key)
	}
	return nil
}

func UserGetCharacterName(userId int) string {
	if user := users.GetByUserId(userId); user != nil {
		return user.Character.Name
	}
	return `Unknown`
}

func UserGetRoomId(userId int) int {
	if user := users.GetByUserId(userId); user != nil {
		return user.Character.RoomId
	}
	return 0
}

func UserHasQuest(userId int, questId string) bool {
	if user := users.GetByUserId(userId); user != nil {
		return user.Character.HasQuest(questId)
	}
	return false
}

func UserGiveQuest(userId int, questId string) {
	commandQueue.QueueQuest(userId, questId)
}

func UserGiveBuff(userId int, buffId int) {
	commandQueue.QueueBuff(userId, 0, buffId)
}

func UserCommand(userId int, cmd string, waitTurns ...int) {
	if len(waitTurns) < 1 {
		waitTurns = append(waitTurns, 0)
	}
	commandQueue.QueueCommand(userId, 0, cmd, waitTurns[0])
}

func UserTrainSkill(userId int, skillName string, skillLevel int) {
	if user := users.GetByUserId(userId); user != nil {

		skillName = strings.ToLower(skillName)
		currentLevel := user.Character.GetSkillLevel(skills.SkillTag(skillName))

		if currentLevel < skillLevel {
			newLevel := user.Character.TrainSkill(skillName, skillLevel)

			skillData := struct {
				SkillName  string
				SkillLevel int
			}{
				SkillName:  skillName,
				SkillLevel: newLevel,
			}
			skillUpTxt, _ := templates.Process("character/skillup", skillData)
			messageQueue.SendUserMessage(userId, skillUpTxt, true)
		}
	}

}

func UserMoveRoom(userId int, destRoomId int) {
	rooms.MoveToRoom(userId, destRoomId)
}

func UserGiveItem(userId int, itemId int) {
	if user := users.GetByUserId(userId); user != nil {
		itm := items.New(itemId)
		if itm.ItemId > 0 {
			user.Character.StoreItem(itm)
			return
		}
	}
}

func UserHasBuffFlag(userId int, buffFlag string) bool {
	if user := users.GetByUserId(userId); user != nil {
		return user.Character.HasBuffFlag(buffs.Flag(buffFlag))
	}
	return false
}

func UserHasItemId(userId int, itemId int) bool {

	if user := users.GetByUserId(userId); user != nil {
		for _, itm := range user.Character.GetAllBackpackItems() {
			if itm.ItemId == itemId {
				return true
			}
		}
	}

	return false
}

func UserGetBackpackItems(userId int) []items.Item {
	if user := users.GetByUserId(userId); user != nil {
		return user.Character.GetAllBackpackItems()
	}
	return []items.Item{}
}
