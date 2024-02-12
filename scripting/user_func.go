package scripting

import (
	"strings"

	"github.com/dop251/goja"
	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/parties"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
)

func setUserFunctions(vm *goja.Runtime) {
	vm.Set(`GetUser`, GetUser)
}

type ScriptUser struct {
	userId     int
	userRecord *users.UserRecord
}

func (u ScriptUser) UserId() int {
	return u.userId
}

func (u ScriptUser) SetTempData(key string, value any) {
	if userValue, ok := value.(ScriptUser); ok { // Don't store pointer to user data.
		userValue.userRecord = nil
		value = userValue
	}
	u.userRecord.SetTempData(key, value)
}

func (u ScriptUser) GetTempData(key string) any {
	if value := u.userRecord.GetTempData(key); value != nil {
		if userValue, ok := value.(ScriptUser); ok { // If it was userdata we need to reload the whole thing in case the user isn't around anymore.
			value = GetUser(userValue.userId)
		}
		return value
	}
	return nil
}

func (u ScriptUser) GetCharacterName() string {
	return u.userRecord.Character.Name
}

func (u ScriptUser) GetRoomId() int {
	return u.userRecord.Character.RoomId
}

func (u ScriptUser) HasQuest(questId string) bool {
	return u.userRecord.Character.HasQuest(questId)
}

func (u ScriptUser) GiveQuest(questId string) {

	// If in a party, give to all party members.
	if party := parties.Get(u.userId); party != nil {
		for _, userId := range party.GetMembers() {
			commandQueue.QueueQuest(userId, questId)
		}
		return
	}

	commandQueue.QueueQuest(u.userId, questId)

}

func (u ScriptUser) GainGold(amt int) {
	u.userRecord.Character.Gold += amt
	if u.userRecord.Character.Gold < 0 {
		u.userRecord.Character.Gold = 0
	}
}

func (u ScriptUser) GiveBuff(buffId int) {
	commandQueue.QueueBuff(u.userId, 0, buffId)
}

func (u ScriptUser) Command(userId int, cmd string, waitTurns ...int) {
	if len(waitTurns) < 1 {
		waitTurns = append(waitTurns, 0)
	}
	commandQueue.QueueCommand(u.userId, 0, cmd, waitTurns[0])
}

func (u ScriptUser) TrainSkill(skillName string, skillLevel int) {

	skillName = strings.ToLower(skillName)
	currentLevel := u.userRecord.Character.GetSkillLevel(skills.SkillTag(skillName))

	if currentLevel < skillLevel {
		newLevel := u.userRecord.Character.TrainSkill(skillName, skillLevel)

		skillData := struct {
			SkillName  string
			SkillLevel int
		}{
			SkillName:  skillName,
			SkillLevel: newLevel,
		}
		skillUpTxt, _ := templates.Process("character/skillup", skillData)
		messageQueue.SendUserMessage(u.userId, skillUpTxt, true)
	}

}

func (u ScriptUser) MoveRoom(destRoomId int) {
	rooms.MoveToRoom(u.userId, destRoomId)
}

func (u ScriptUser) GiveItem(itemId any) {

	if id, ok := itemId.(int); ok {
		itm := items.New(id)
		if itm.ItemId > 0 {
			u.userRecord.Character.StoreItem(itm)
		}
		return
	}

	if itmObj, ok := itemId.(items.Item); ok {
		u.userRecord.Character.StoreItem(itmObj)
		return
	}

}

func (u ScriptUser) HasBuffFlag(buffFlag string) bool {
	return u.userRecord.Character.HasBuffFlag(buffs.Flag(buffFlag))
}

func (u ScriptUser) HasItemId(itemId int) bool {
	for _, itm := range u.userRecord.Character.GetAllBackpackItems() {
		if itm.ItemId == itemId {
			return true
		}
	}
	return false
}

func (u ScriptUser) GetBackpackItems() []items.Item {
	return u.userRecord.Character.GetAllBackpackItems()
}

func (u ScriptUser) GetAlignment() int {
	return int(u.userRecord.Character.Alignment)
}

func (u ScriptUser) GetAlignmentName() string {
	return u.userRecord.Character.AlignmentName()
}

func (u ScriptUser) SetAlignment(newAlignment int) {
	if newAlignment < -100 {
		newAlignment = -100
	} else if newAlignment > 100 {
		newAlignment = 100
	}

	u.userRecord.Character.Alignment = int8(newAlignment)
}

// ////////////////////////////////////////////////////////
//
// # These functions get exported to the scripting engine
//
// ////////////////////////////////////////////////////////
func GetUser(userId int) *ScriptUser {
	if user := users.GetByUserId(userId); user != nil {
		return &ScriptUser{userId, user}
	}
	return nil
}
