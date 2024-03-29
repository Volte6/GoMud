package scripting

import (
	"strings"

	"github.com/dop251/goja"
	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/parties"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
)

func setActorFunctions(vm *goja.Runtime) {
	vm.Set(`GetUser`, GetUser)
	vm.Set(`GetMob`, GetMob)
}

type ScriptActor struct {
	userId          int
	mobInstanceId   int
	userRecord      *users.UserRecord
	mobRecord       *mobs.Mob
	characterRecord *characters.Character // Lets us bypass the user/mob check in many cases
}

func (a ScriptActor) UserId() int {
	return a.userId
}

func (a ScriptActor) InstanceId() int {
	return a.mobInstanceId
}

func (a ScriptActor) MobTypeId() int {
	if a.mobRecord != nil {
		return int(a.mobRecord.MobId)
	}
	return 0
}

func (a ScriptActor) SetTempData(key string, value any) {

	if a.userRecord != nil {
		if userValue, ok := value.(ScriptActor); ok { // Don't store pointer to user data.
			userValue.userRecord = nil
			value = userValue
		}
		a.userRecord.SetTempData(key, value)
		return
	}

	if a.mobRecord != nil {
		if userValue, ok := value.(ScriptActor); ok { // Don't store pointer to user data.
			userValue.mobRecord = nil
			value = userValue
		}
		a.mobRecord.SetTempData(key, value)
		return
	}
}

func (a ScriptActor) GetTempData(key string) any {

	if a.userRecord != nil {
		if value := a.userRecord.GetTempData(key); value != nil {
			if userValue, ok := value.(ScriptActor); ok { // If it was userdata we need to reload the whole thing in case the user isn't around anymore.
				value = GetActor(userValue.userId, 0)
			}
			return value
		}
	} else if a.mobRecord != nil {
		if value := a.mobRecord.GetTempData(key); value != nil {
			if mobValue, ok := value.(ScriptActor); ok { // If it was userdata we need to reload the whole thing in case the user isn't around anymore.
				value = GetActor(0, mobValue.mobInstanceId)
			}
			return value
		}
	}
	return nil
}

func (a ScriptActor) GetCharacterName(wrapInTags ...bool) string {

	if len(wrapInTags) > 0 && wrapInTags[0] {
		if a.userRecord != nil {
			return `<ansi fg="username">` + a.characterRecord.Name + `</ansi>`
		} else if a.mobRecord != nil {
			return `<ansi fg="mobname">` + a.characterRecord.Name + `</ansi>`
		}
	}

	return a.characterRecord.Name
}

func (a ScriptActor) GetRoomId() int {
	return a.characterRecord.RoomId
}

func (a ScriptActor) HasQuest(questId string) bool {
	return a.characterRecord.HasQuest(questId)
}

func (a ScriptActor) GiveQuest(questId string) {

	if a.userRecord != nil {
		// If in a party, give to all party members.
		if party := parties.Get(a.userId); party != nil {
			for _, userId := range party.GetMembers() {
				commandQueue.QueueQuest(userId, questId)
			}
			return
		}
	}
	a.characterRecord.GiveQuestToken(questId)

}

func (a ScriptActor) AddGold(amt int, bankAmt ...int) {
	a.characterRecord.Gold += amt
	if a.characterRecord.Gold < 0 {
		a.characterRecord.Gold = 0
	}
	if len(bankAmt) > 0 {
		a.characterRecord.Bank += bankAmt[0]
		if a.characterRecord.Bank < 0 {
			a.characterRecord.Bank = 0
		}
	}
}

func (a ScriptActor) AddHealth(amt int) int {
	return a.characterRecord.ApplyHealthChange(amt)
}

func (a ScriptActor) Command(cmd string, waitTurns ...int) {
	if len(waitTurns) < 1 {
		waitTurns = append(waitTurns, 0)
	}
	commandQueue.QueueCommand(a.userId, a.mobInstanceId, cmd, waitTurns[0])
}

func (a ScriptActor) TrainSkill(skillName string, skillLevel int) {

	skillName = strings.ToLower(skillName)
	currentLevel := a.characterRecord.GetSkillLevel(skills.SkillTag(skillName))

	if currentLevel < skillLevel {
		newLevel := a.characterRecord.TrainSkill(skillName, skillLevel)

		if a.userRecord != nil {
			skillData := struct {
				SkillName  string
				SkillLevel int
			}{
				SkillName:  skillName,
				SkillLevel: newLevel,
			}
			skillUpTxt, _ := templates.Process("character/skillup", skillData)
			messageQueue.SendUserMessage(a.userId, skillUpTxt, true)
		}

	}

}

func (a ScriptActor) MoveRoom(destRoomId int) {

	if a.userRecord != nil {

		rooms.MoveToRoom(a.userId, destRoomId)

	} else if a.mobRecord != nil {

		if mobRoom := rooms.LoadRoom(a.characterRecord.RoomId); mobRoom != nil {
			if destRoom := rooms.LoadRoom(destRoomId); destRoom != nil {
				mobRoom.RemoveMob(a.mobInstanceId)
				destRoom.AddMob(a.mobInstanceId)
			}
		}

	}
}

func (a ScriptActor) GiveItem(itemId any) {

	if id, ok := itemId.(int); ok {
		itm := items.New(id)
		if itm.ItemId > 0 {
			a.characterRecord.StoreItem(itm)
		}
		return
	}

	if itmObj, ok := itemId.(items.Item); ok {
		a.characterRecord.StoreItem(itmObj)
		return
	}

}

func (a ScriptActor) HasBuff(buffId int) bool {
	return a.characterRecord.HasBuff(buffId)
}

func (a ScriptActor) GiveBuff(buffId int) {
	commandQueue.QueueBuff(a.userId, a.mobInstanceId, buffId)
}

func (a ScriptActor) HasBuffFlag(buffFlag string) bool {
	return a.characterRecord.HasBuffFlag(buffs.Flag(buffFlag))
}

func (a ScriptActor) CancelBuffWithFlag(buffFlag string) bool {
	return a.characterRecord.CancelBuffsWithFlag(buffs.Flag(buffFlag))
}

func (a ScriptActor) ExpireBuff(buffId int) {
	a.characterRecord.Buffs.CancelBuffId(buffId)
}

func (a ScriptActor) RemoveBuff(buffId int) {
	a.characterRecord.Buffs.RemoveBuff(buffId * -1)
}

func (a ScriptActor) HasItemId(itemId int) bool {
	for _, itm := range a.characterRecord.GetAllBackpackItems() {
		if itm.ItemId == itemId {
			return true
		}
	}
	return false
}

func (a ScriptActor) GetBackpackItems() []items.Item {
	return a.characterRecord.GetAllBackpackItems()
}

func (a ScriptActor) GetAlignment() int {
	return int(a.characterRecord.Alignment)
}

func (a ScriptActor) GetAlignmentName() string {
	return a.characterRecord.AlignmentName()
}

func (a ScriptActor) ChangeAlignment(alignmentChange int) {
	newAlignment := int(a.characterRecord.Alignment) + alignmentChange
	if newAlignment < -100 {
		newAlignment = -100
	} else if newAlignment > 100 {
		newAlignment = 100
	}

	a.characterRecord.Alignment = int8(newAlignment)
}

// ////////////////////////////////////////////////////////
//
// Functions only really useful for mobs
//
// ////////////////////////////////////////////////////////

// Returns true if a mob is charmed by/friendly to a player.
// If userId is ommitted, it will return true if the mob is charmed by any player.
func (a ScriptActor) IsCharmed(userId ...int) bool {
	if len(userId) < 1 {
		return a.characterRecord.IsCharmed()
	}
	return a.characterRecord.IsCharmed(userId[0])
}

func (a ScriptActor) CharmSet(userId int, charmRounds int, onRevertCommand ...string) {
	if len(onRevertCommand) < 1 {
		onRevertCommand = append(onRevertCommand, ``)
	}
	a.characterRecord.Charm(userId, charmRounds, onRevertCommand[0])
}

func (a ScriptActor) CharmRemove() {
	if a.characterRecord.Charmed == nil {
		return
	}
	a.characterRecord.RemoveCharm()
}

func (a ScriptActor) CharmExpire() {
	a.characterRecord.Charmed.Expire()
}

func (a ScriptActor) getScript() string {
	if a.mobRecord != nil {
		return a.mobRecord.GetScript()
	}
	return ""
}

func (a ScriptActor) getScriptTag() string {
	if a.mobRecord != nil {
		return a.mobRecord.ScriptTag
	}
	return ""
}

// ////////////////////////////////////////////////////////
//
// # These functions get exported to the scripting engine
//
// ////////////////////////////////////////////////////////
func GetActor(userId int, mobInstanceId int) *ScriptActor {

	if userId > 0 {
		if user := users.GetByUserId(userId); user != nil {
			return &ScriptActor{
				userId:          userId,
				userRecord:      user,
				characterRecord: user.Character,
			}
		}
	} else if mobInstanceId > 0 {
		if mob := mobs.GetInstance(mobInstanceId); mob != nil {
			return &ScriptActor{
				mobInstanceId:   mobInstanceId,
				mobRecord:       mob,
				characterRecord: &mob.Character,
			}
		}
	}

	return nil
}

func GetUser(userId int) *ScriptActor {
	return GetActor(userId, 0)
}

func GetMob(mobInstanceId int) *ScriptActor {
	return GetActor(0, mobInstanceId)
}
