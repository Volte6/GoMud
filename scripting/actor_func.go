package scripting

import (
	"strings"

	"github.com/dop251/goja"
	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/parties"
	"github.com/volte6/mud/races"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
)

func setActorFunctions(vm *goja.Runtime) {
	vm.Set(`GetUser`, GetUser)
	vm.Set(`GetMob`, GetMob)
	vm.Set(`ActorNames`, ActorNames)
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

func (a ScriptActor) GetRace() string {
	return a.characterRecord.Race()
}

func (a ScriptActor) GetSize() string {
	if r := races.GetRace(a.characterRecord.RaceId); r != nil {
		return string(r.Size)
	}
	return string(races.Medium)
}

func (a ScriptActor) GetLevel() int {
	return a.characterRecord.Level
}

func (a ScriptActor) GetStat(statName string) int {

	statName = strings.ToLower(statName)

	if strings.HasPrefix(statName, "st") {
		return a.characterRecord.Stats.Strength.ValueAdj
	}

	if strings.HasPrefix(statName, "sp") {
		return a.characterRecord.Stats.Speed.ValueAdj
	}

	if strings.HasPrefix(statName, "sm") {
		return a.characterRecord.Stats.Smarts.ValueAdj
	}

	if strings.HasPrefix(statName, "vi") {
		return a.characterRecord.Stats.Vitality.ValueAdj
	}

	if strings.HasPrefix(statName, "my") {
		return a.characterRecord.Stats.Mysticism.ValueAdj
	}

	if strings.HasPrefix(statName, "pe") {
		return a.characterRecord.Stats.Perception.ValueAdj
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

func (a ScriptActor) SetMiscCharacterData(key string, value any) {

	if _, ok := value.(ScriptActor); ok { // Don't store actor data.
		return
	}
	a.characterRecord.SetMiscData(key, value)
}

func (a ScriptActor) GetMiscCharacterData(key string) any {
	if value := a.characterRecord.GetMiscData(key); value != nil {
		return value
	}
	return nil
}

func (a ScriptActor) GetMiscCharacterDataKeys(prefixMatches ...string) []string {
	return a.characterRecord.GetMiscDataKeys(prefixMatches...)
}

func (a ScriptActor) GetCharacterName(wrapInTags bool) string {

	if wrapInTags {
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
		} else {
			commandQueue.QueueQuest(a.userId, questId)
		}
	}
	//a.characterRecord.GiveQuestToken(questId)

}

func (a ScriptActor) GetPartyMembers() []ScriptActor {

	partyMembers := []ScriptActor{}
	partyUserId := 0

	if a.userRecord == nil {
		if a.mobRecord.Character.Charmed == nil {
			return partyMembers
		}

		partyUserId = a.mobRecord.Character.Charmed.UserId
	} else {
		partyUserId = a.userId
	}

	if partyUserId < 1 {
		return partyMembers
	}

	// If in a party, give to all party members.
	if party := parties.Get(partyUserId); party != nil {
		for _, userId := range party.GetMembers() {

			if a := GetActor(userId, 0); a != nil {
				partyMembers = append(partyMembers, *a)
			}

		}
	}

	mobPartyMembers := []ScriptActor{}

	for _, char := range partyMembers {
		for _, mobInstId := range char.characterRecord.GetCharmIds() {
			if a := GetActor(0, mobInstId); a != nil {
				mobPartyMembers = append(mobPartyMembers, *a)
			}
		}
	}

	return append(partyMembers, mobPartyMembers...)
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

func (a ScriptActor) AddMana(amt int) int {
	return a.characterRecord.ApplyManaChange(amt)
}

func (a ScriptActor) Command(cmd string, waitTurns ...int) {
	if len(waitTurns) < 1 {
		waitTurns = append(waitTurns, 0)
	}
	commandQueue.QueueCommand(a.userId, a.mobInstanceId, cmd, waitTurns[0])
}

func (a ScriptActor) TrainSkill(skillName string, skillLevel int) bool {

	if a.userRecord == nil {
		return false
	}

	skillName = strings.ToLower(skillName)
	currentLevel := a.characterRecord.GetSkillLevel(skills.SkillTag(skillName))

	if currentLevel < skillLevel {
		newLevel := a.characterRecord.TrainSkill(skillName, skillLevel)

		skillData := struct {
			SkillName  string
			SkillLevel int
		}{
			SkillName:  skillName,
			SkillLevel: newLevel,
		}
		skillUpTxt, _ := templates.Process("character/skillup", skillData)
		messageQueue.SendUserMessage(a.userId, skillUpTxt, true)

		return true

	}
	return false
}

func (a ScriptActor) GetSkillLevel(skillName string) int {
	return a.characterRecord.GetSkillLevel(skills.SkillTag(skillName))
}

func (a ScriptActor) MoveRoom(destRoomId int, leaveCharmedMobs ...bool) {

	if a.userRecord != nil {

		rmNow := rooms.LoadRoom(a.characterRecord.RoomId)

		if rmNext := rooms.LoadRoom(destRoomId); rmNext != nil {
			rooms.MoveToRoom(a.userId, destRoomId)

			if len(leaveCharmedMobs) < 1 || !leaveCharmedMobs[0] {
				for _, mobInstId := range a.characterRecord.GetCharmIds() {
					rmNow.RemoveMob(mobInstId)
					rmNext.AddMob(mobInstId)
				}
			}
		}

	} else if a.mobRecord != nil {

		if mobRoom := rooms.LoadRoom(a.characterRecord.RoomId); mobRoom != nil {
			if destRoom := rooms.LoadRoom(destRoomId); destRoom != nil {
				mobRoom.RemoveMob(a.mobInstanceId)
				destRoom.AddMob(a.mobInstanceId)
			}
		}

	}
}

func (a ScriptActor) UpdateItem(itm ScriptItem) {
	a.userRecord.Character.UpdateItem(itm.originalItem, *itm.itemRecord)
}

func (a ScriptActor) GiveItem(itm ScriptItem) {
	if a.characterRecord.StoreItem(*itm.itemRecord) {
		if a.userId > 0 {
			TryItemScriptEvent(`onGive`, *itm.itemRecord, a.userId, commandQueue)
		}
	}
}

func (a ScriptActor) TakeItem(itm ScriptItem) {
	if a.characterRecord.RemoveItem(*itm.itemRecord) {
		if a.userId > 0 {
			TryItemScriptEvent(`onLost`, *itm.itemRecord, a.userId, commandQueue)
		}
	}
}

func (a ScriptActor) IsTameable() bool {
	if a.mobRecord == nil {
		return false
	}
	return a.mobRecord.IsTameable()
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

	found := false

	for _, buffId := range a.characterRecord.Buffs.GetBuffIdsWithFlag(buffs.Flag(strings.ToLower(buffFlag))) {
		found = found || a.RemoveBuff(buffId)
	}

	return found
}

// Remove a buff silently
func (a ScriptActor) RemoveBuff(buffId int) bool {

	if !configs.GetConfig().AllowItemBuffRemoval {
		buffList := a.characterRecord.GetBuffs(buffId)
		if len(buffList) > 0 {
			if buffList[0].ItemBuff {
				return false
			}
		}
	}

	return a.characterRecord.Buffs.RemoveBuff(buffId)

}

func (a ScriptActor) HasItemId(itemId int, excludeWorn ...bool) bool {
	for _, itm := range a.characterRecord.GetAllBackpackItems() {
		if itm.ItemId == itemId {
			return true
		}
	}
	if len(excludeWorn) == 0 || !excludeWorn[0] {
		for _, itm := range a.characterRecord.GetAllWornItems() {
			if itm.ItemId == itemId {
				return true
			}
		}
	}
	return false
}

func (a ScriptActor) GetBackpackItems() []ScriptItem {
	itms := make([]ScriptItem, 0, 5)
	for _, item := range a.characterRecord.GetAllBackpackItems() {
		itms = append(itms, newScriptItem(item))
	}
	return itms
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

func (a ScriptActor) HasSpell(spellId string) bool {
	return a.characterRecord.HasSpell(spellId)
}

func (a ScriptActor) LearnSpell(spellId string) bool {
	return a.characterRecord.LearnSpell(spellId)
}

func (a ScriptActor) IsAggro(actor ScriptActor) bool {
	return a.characterRecord.IsAggro(actor.UserId(), actor.InstanceId())
}

func (a ScriptActor) GetMobKills(mobId int) int {
	return a.characterRecord.KD.GetMobKills(mobId)
}

func (a ScriptActor) GetRaceKills(race string) int {
	return a.characterRecord.KD.GetRaceKills(race)
}

func (a ScriptActor) GetHealth() int {
	return a.characterRecord.Health
}

func (a ScriptActor) GetHealthMax() int {
	return a.characterRecord.HealthMax.Value
}

func (a ScriptActor) GetHealthPct() float64 {
	return float64(a.characterRecord.Health) / float64(a.characterRecord.HealthMax.Value)
}

func (a ScriptActor) GetMana() int {
	return a.characterRecord.Mana
}

func (a ScriptActor) GetManaMax() int {
	return a.characterRecord.ManaMax.Value
}

func (a ScriptActor) GetManaPct() float64 {
	return float64(a.characterRecord.Mana) / float64(a.characterRecord.ManaMax.Value)
}

func (a ScriptActor) SetAdjective(adj string, addIt bool) {
	a.characterRecord.SetAdjective(adj, addIt)
}

func (a ScriptActor) GetCharmCount() int {
	return len(a.characterRecord.GetCharmIds())
}

func (a ScriptActor) GiveTrainingPoints(ct int) {
	if ct < 1 {
		return
	}
	a.characterRecord.TrainingPoints += ct
}

func (a ScriptActor) GiveStatPoints(ct int) {
	if ct < 1 {
		return
	}
	a.characterRecord.StatPoints += ct
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

	// If the player is in a party, add the mob to their party
	if a.mobInstanceId < 1 {
		return
	}

	if len(onRevertCommand) < 1 {
		onRevertCommand = append(onRevertCommand, ``)
	}
	a.characterRecord.Charm(userId, charmRounds, onRevertCommand[0])

	if user := users.GetByUserId(userId); user != nil {
		user.Character.TrackCharmed(a.mobInstanceId, true)
	}

}

func (a ScriptActor) CharmRemove() {
	if a.characterRecord.Charmed == nil {
		return
	}
	charmUserId := a.characterRecord.RemoveCharm()

	if user := users.GetByUserId(charmUserId); user != nil {
		user.Character.TrackCharmed(a.mobInstanceId, false)
	}
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

func ActorNames(actorList []*ScriptActor) string {

	sBuilder := strings.Builder{}
	listSize := len(actorList)

	for i := 0; i < listSize; i++ {

		sBuilder.WriteString(actorList[i].GetCharacterName(true))

		if i < listSize-2 {
			sBuilder.WriteString(`, `)
		} else if i == listSize-2 {
			sBuilder.WriteString(`and `)
		}
	}

	return sBuilder.String()
}
