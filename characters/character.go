package characters

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"log/slog"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/quests"
	"github.com/volte6/mud/races"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/stats"
	"github.com/volte6/mud/util"
	//
)

var (
	startingRace     = 0
	startingHealth   = 10
	startingMana     = 10
	startingRoomId   = -1
	startingZone     = `Nowhere`
	defaultName      = `Nameless`
	descriptionCache = map[string]string{} // key is a hash, value is the description
)

type NameRenderFlag uint8

const (
	AlignmentMinimum int8 = -100
	AlignmentNeutral int8 = 0
	AlignmentMaximum int8 = 100

	RenderHealth NameRenderFlag = iota
	RenderAggro
	RenderShortAdjectives
)

type Character struct {
	Name            string            // The name of the character
	Description     string            // A description of the character.
	Adjectives      []string          // Decorative text for the name of the character (e.g. "sleeping", "dead", "wounded")
	RoomId          int               // The room id the character is in.
	Zone            string            // The zone the character is in. The folder the room can be located in too.
	RaceId          int               // Character race
	Stats           stats.Statistics  // Character stats
	Level           int               // The level of the character
	Experience      int               // The experience of the character
	TrainingPoints  int               // The number of training points the character has
	StatPoints      int               // The number of skill points the character has
	Health          int               // The health of the character
	Mana            int               // The mana of the character
	ActionPoints    int               // The resevoir of action points the character has to spend on movement etc.
	Alignment       int8              // The alignment of the character
	Gold            int               // The gold the character is holding
	Bank            int               // The gold the character has in the bank
	SpellBook       map[string]int    `yaml:"spellbook,omitempty"` // The spells the character has learned
	Charmed         *CharmInfo        `yaml:"-"`                   // If they are charmed, this is the info
	CharmedMobs     []int             `yaml:"-"`                   // If they have charmed anyone, this is the list of mob instance ids
	Items           []items.Item      // The items the character is holding
	Buffs           buffs.Buffs       `yaml:"buffs,omitempty"` // The buffs the character has active
	Equipment       Worn              // The equipment the character is wearing
	Energy          int               `yaml:"energy,omitempty"`        // The energy the character has
	TNLScale        float32           `yaml:"-"`                       // The experience scale of the character. Don't write to yaml since is dynamically calculated.
	HealthMax       stats.StatInfo    `yaml:"-"`                       // The maximum health of the character. Don't write to yaml since is dynamically calculated.
	ManaMax         stats.StatInfo    `yaml:"-"`                       // The maximum mana of the character. Don't write to yaml since is dynamically calculated.
	ActionPointsMax stats.StatInfo    `yaml:"-"`                       // The maximum actions of character. Don't write to yaml since is dynamically calculated.
	Aggro           *Aggro            `yaml:"-"`                       // Dont' store this. If they leave they break their aggro
	Skills          map[string]int    `yaml:"skills,omitempty"`        // The skills the character has, and what level they are at
	Cooldowns       Cooldowns         `yaml:"cooldowns,omitempty"`     // How many rounds until it is cooled down
	Settings        map[string]string `yaml:"settings,omitempty"`      // custom setting tracking, used for anything.
	QuestProgress   map[int]string    `yaml:"questprogress,omitempty"` // quest progress tracking
	KeyRing         map[string]string `yaml:"keyring,omitempty"`       // key is the lock id, value is the sequence
	KD              KDStats           `yaml:"kd,omitempty"`            // Kill/Death stats
	MiscData        map[string]any    `yaml:"miscdata,omitempty"`      // Any random other data that needs to be stored
	roomHistory     []int             // A stack FILO of the last X rooms the character has been in
	followers       []int             `yaml:"-"` // everyone following this user
}

func New() *Character {
	return &Character{
		//Name:   defaultName,
		Adjectives: []string{},
		RoomId:     startingRoomId,
		Zone:       startingZone,
		RaceId:     startingRace,
		Stats: stats.Statistics{
			Strength:   stats.StatInfo{Base: 1},
			Speed:      stats.StatInfo{Base: 1},
			Smarts:     stats.StatInfo{Base: 1},
			Vitality:   stats.StatInfo{Base: 1},
			Mysticism:  stats.StatInfo{Base: 1},
			Perception: stats.StatInfo{Base: 1},
		},
		Level:          1,
		Experience:     1,
		TrainingPoints: 0,
		StatPoints:     0,
		TNLScale:       1.0,
		Health:         startingHealth,
		HealthMax:      stats.StatInfo{Base: 1},
		Mana:           startingMana,
		ManaMax:        stats.StatInfo{Base: 1},
		Skills:         make(map[string]int),
		Gold:           25,
		Bank:           100,
		SpellBook:      make(map[string]int),
		CharmedMobs:    []int{},
		Items:          []items.Item{},
		Buffs:          buffs.New(),
		Equipment:      Worn{},
		MiscData:       make(map[string]any),
		roomHistory:    make([]int, 0, 10),
		KeyRing:        make(map[string]string),
	}
}

// returns description unless description is a hash
// which points to another description location.
func (c *Character) GetDescription() string {

	if !strings.HasPrefix(c.Description, `h:`) {
		return c.Description
	}
	hash := strings.TrimPrefix(c.Description, `h:`)
	return descriptionCache[hash]
}

func (c *Character) DeductActionPoints(amount int) bool {

	if c.ActionPoints < amount {
		return false
	}
	c.ActionPoints -= 10
	if c.ActionPoints < 0 {
		c.ActionPoints = 0
	}
	return true
}

func (c *Character) SetMiscData(key string, value any) {

	if c.MiscData == nil {
		c.MiscData = make(map[string]any)
	}

	if value == nil {
		delete(c.MiscData, key)
		return
	}
	c.MiscData[key] = value
}

func (c *Character) GetMiscData(key string) any {

	if c.MiscData == nil {
		c.MiscData = make(map[string]any)
	}

	if value, ok := c.MiscData[key]; ok {
		return value
	}
	return nil
}

func (c *Character) GetMiscDataKeys(prefixMatch ...string) []string {

	if c.MiscData == nil {
		c.MiscData = make(map[string]any)
	}

	allKeys := []string{}
	for key, _ := range c.MiscData {
		allKeys = append(allKeys, key)
	}

	if len(prefixMatch) == 0 {
		return allKeys
	}

	retKeys := []string{}
	for _, prefix := range prefixMatch {
		for _, key := range allKeys {
			if finalKey, ok := strings.CutPrefix(key, prefix); ok {
				retKeys = append(retKeys, finalKey)
			}
		}
	}

	return retKeys
}

func (c *Character) FindKeyInBackpack(lockId string) (items.Item, bool) {

	for _, itm := range c.GetAllBackpackItems() {
		itmSpec := itm.GetSpec()
		if itmSpec.Type != items.Key {
			continue
		}

		if itmSpec.KeyLockId == lockId {
			return itm, true
		}
	}

	return items.Item{}, false
}

func (c *Character) HasKey(lockId string, difficulty int) (hasKey bool, hasSequence bool) {

	sequence := util.GetLockSequence(lockId, difficulty, string(configs.GetConfig().Seed))

	// Check whether they ahve a key for this lock
	return c.GetKey(`key-`+lockId) != ``, c.GetKey(lockId) == sequence
}

func (c *Character) KeyCount() int {
	if c.KeyRing == nil {
		c.KeyRing = make(map[string]string)
	}
	return len(c.KeyRing)
}

func (c *Character) GetKey(lockId string) string {
	if c.KeyRing == nil {
		c.KeyRing = make(map[string]string)
	}
	return c.KeyRing[strings.ToLower(lockId)]
}

func (c *Character) SetKey(lockId string, sequence string) {
	if c.KeyRing == nil {
		c.KeyRing = make(map[string]string)
	}
	if len(sequence) == 0 {
		delete(c.KeyRing, strings.ToLower(lockId))
	} else {
		c.KeyRing[strings.ToLower(lockId)] = strings.ToUpper(sequence)
	}
}

// This should only be used for mobs.
// Not players
func (c *Character) CacheDescription() {
	// Hash the descriptions and store centrally.
	// This saves a lot of memory because many descriptions are duplicates
	hash := util.Hash(c.Description)
	if _, ok := descriptionCache[hash]; !ok {
		descriptionCache[hash] = c.Description
	}
	c.Description = fmt.Sprintf(`h:%s`, hash)
}

func (c *Character) GetDefaultDiceRoll() (attacks int, dCount int, dSides int, bonus int, buffOnCrit []int) {
	// default racial
	raceInfo := races.GetRace(c.RaceId)

	attacks = raceInfo.Damage.Attacks
	dCount = raceInfo.Damage.DiceCount
	dSides = raceInfo.Damage.SideCount
	bonus = raceInfo.Damage.BonusDamage
	buffOnCrit = raceInfo.Damage.CritBuffIds

	dCount += int(math.Floor((float64(c.Stats.Speed.Value) / 50)))
	dSides += int(math.Floor((float64(c.Stats.Strength.Value) / 12)))
	bonus += int(math.Floor((float64(c.Stats.Perception.Value) / 25)))

	if dCount < 1 {
		dCount = 1
	}
	if dSides < 2 {
		dSides = 2
	}

	return attacks, dCount, dSides, bonus, buffOnCrit
}

func (c *Character) GetSpells() map[string]int {
	ret := make(map[string]int)
	for sName, sCasts := range c.SpellBook {
		ret[sName] = sCasts
	}
	return ret
}

func (c *Character) HasSpell(spellName string) bool {
	if intVal, ok := c.SpellBook[spellName]; ok {
		return intVal > 0
	}
	return false
}

func (c *Character) DisableSpell(spellName string) bool {
	if intVal, ok := c.SpellBook[spellName]; ok {
		if intVal > 0 {
			c.SpellBook[spellName] = intVal * -1
		}
	}
	return false
}

func (c *Character) EnableSpell(spellName string) bool {
	if intVal, ok := c.SpellBook[spellName]; ok {
		if intVal < 0 {
			c.SpellBook[spellName] = intVal * -1
		}
	}
	return false
}

func (c *Character) TrackSpellCast(spellName string) bool {
	if intVal, ok := c.SpellBook[spellName]; ok {
		if intVal > 0 {
			intVal++
			c.SpellBook[spellName] = intVal
		}
	}
	return false
}

func (c *Character) LearnSpell(spellName string) bool {
	if _, ok := c.SpellBook[spellName]; !ok {
		c.SpellBook[spellName] = 1
		return true
	}
	return false
}

func (c *Character) GrantXP(xp int) (actualXP int, xpScale int) {

	if xp == 0 {
		return 0, 100
	}

	xpScale = c.Buffs.StatMod("xpscale") + 100

	if xpScale == 100 {
		actualXP = xp
	} else {

		scaleFloat := float64(xpScale) / 100
		if scaleFloat < 1 {
			scaleFloat = 1
		}

		actualXP = int(float64(xp) * scaleFloat)
	}

	c.Experience += actualXP

	slog.Info(`GrantXP()`, `username`, c.Name, `xp`, xp, `xpscale`, xpScale, `actualXP`, actualXP)

	return actualXP, xpScale
}

func (c *Character) TrackCharmed(mobId int, add bool) {
	for _, mobInstanceId := range c.CharmedMobs {
		if mobInstanceId == mobId {
			if !add {
				c.CharmedMobs = append(c.CharmedMobs[:mobInstanceId], c.CharmedMobs[mobInstanceId+1:]...)
			}
			return
		}
	}
	c.CharmedMobs = append(c.CharmedMobs, mobId)
}

func (c *Character) GetCharmIds() []int {
	return append([]int{}, c.CharmedMobs...)
}

func (c *Character) Charm(userId int, rounds int, expireCommand string) {
	c.SetAdjective(`charmed`, true)
	c.Charmed = NewCharm(userId, rounds, expireCommand)
	if c.Aggro != nil && c.Aggro.UserId == userId {
		c.Aggro = nil
	}
}

func (c *Character) KnowsFirstAid() bool {
	if r := races.GetRace(c.RaceId); r != nil {
		return r.KnowsFirstAid
	}
	return false
}

func (c *Character) GetCharmedUserId() int {
	if c.Charmed != nil {
		return c.Charmed.UserId
	}
	return 0
}

func (c *Character) IsCharmed(userId ...int) bool {

	if c.Charmed == nil {
		return false
	}

	if len(userId) == 0 {
		return c.Charmed != nil
	}

	for _, uId := range userId {
		if c.Charmed.UserId == uId {
			return true
		}
	}

	return false
}

// Returns userId of whoever had charmed them
func (c *Character) RemoveCharm() int {
	charmUserId := 0
	c.SetAdjective(`charmed`, false)
	if c.Charmed != nil {
		charmUserId = c.Charmed.UserId
		c.Charmed = nil
	}
	return charmUserId
}

func (c *Character) GetRandomItem() (items.Item, bool) {
	if len(c.Items) == 0 {
		return items.Item{}, false
	}
	return c.Items[util.Rand(len(c.Items))], true
}

func (c *Character) AddFollower(uId int) {
	c.followers = append(c.followers, uId)
}

// USERNAME appears to be <BLANK>
func (c *Character) GetHealthAppearance() string {

	className := util.HealthClass(c.Health, c.HealthMax.Value)
	pct := int(float64(c.Health) / float64(c.HealthMax.Value) * 100)

	if pct < 15 {
		return fmt.Sprintf(`<ansi fg="username">%s</ansi> looks like they're <ansi fg="%s">about to die!</ansi>`, c.Name, className)
	}

	if pct < 50 {
		return fmt.Sprintf(`<ansi fg="username">%s</ansi> looks to be in <ansi fg="%s">pretty bad shape.</ansi>`, c.Name, className)
	}

	if pct < 80 {
		return fmt.Sprintf(`<ansi fg="username">%s</ansi> has some <ansi fg="%s">cuts and bruises.</ansi>`, c.Name, className)
	}

	if pct < 100 {
		return fmt.Sprintf(`<ansi fg="username">%s</ansi> has <ansi fg="%s">a few scratches.</ansi>`, c.Name, className)
	}

	return fmt.Sprintf(`<ansi fg="username">%s</ansi> is in <ansi fg="%s">perfect health.</ansi>`, c.Name, className)
}

func (c *Character) GetBackpackCapacity() int {
	return int(math.Ceil(float64(c.Stats.Strength.Value)/3)) + 3
}

func (c *Character) GetFollowers() []int {
	return append([]int{}, c.followers...)
}

func (c *Character) GetAllSkillRanks() map[string]int {
	retMap := make(map[string]int)
	for skillName, skillLevel := range c.Skills {
		retMap[skillName] = skillLevel
	}
	return retMap
}

// Returns an integer representing a % damage reduction
func (c *Character) GetDefense() int {

	reduction := c.Equipment.Weapon.GetDefense() +
		c.Equipment.Offhand.GetDefense() +
		c.Equipment.Head.GetDefense() +
		c.Equipment.Neck.GetDefense() +
		c.Equipment.Body.GetDefense() +
		c.Equipment.Belt.GetDefense() +
		c.Equipment.Gloves.GetDefense() +
		c.Equipment.Ring.GetDefense() +
		c.Equipment.Legs.GetDefense() +
		c.Equipment.Feet.GetDefense()

	//reduction = int(float64(reduction) / 9)

	// If wearing an offhand item like a shield, defense gets a 50% boost
	// Holdables are not considered "shield" type items.
	if c.Equipment.Offhand.ItemId != 0 && c.Equipment.Offhand.GetSpec().Type != items.Weapon && c.Equipment.Offhand.GetSpec().Type != items.Holdable {
		reduction = int(float64(reduction) * 1.5)
	}

	if reduction > 100 {
		reduction = 100
	}

	return reduction
}

func (c *Character) GetMobName(viewingUserId int, renderFlags ...NameRenderFlag) FormattedName {
	return c.getFormattedName(viewingUserId, `mobname`, renderFlags...)
}

func (c *Character) GetPlayerName(viewingUserId int, renderFlags ...NameRenderFlag) FormattedName {
	return c.getFormattedName(viewingUserId, `username`, renderFlags...)
}

func (c *Character) getFormattedName(viewingUserId int, uType string, renderFlags ...NameRenderFlag) FormattedName {
	f := FormattedName{
		Name:       c.Name,
		Type:       uType,
		Adjectives: make([]string, 0, len(c.Adjectives)),
	}

	includeHealth := false
	for _, flag := range renderFlags {
		if flag == RenderHealth {
			includeHealth = true
		} else if flag == RenderShortAdjectives {
			f.UseShortAdjectives = true
		}
	}

	if includeHealth {
		if c.Health < 1 {
			f.Adjectives = append(f.Adjectives, `downed`)
		} else {
			pctHealth := int(math.Ceil(float64(c.Health) / float64(c.HealthMax.Value) * 100))
			f.Adjectives = append(f.Adjectives, strconv.Itoa(pctHealth)+`%`)
		}
	}

	f.Adjectives = append(f.Adjectives, c.Adjectives...)

	if c.HasBuffFlag(buffs.EmitsLight) {
		f.Adjectives = append(f.Adjectives, `lit`)
	}

	if c.HasBuffFlag(buffs.Hidden) {
		f.Adjectives = append(f.Adjectives, `hidden`)
	}

	if c.HasBuffFlag(buffs.Poison) {
		f.Adjectives = append(f.Adjectives, `poisoned`)
	}

	if c.Health < 1 {
		f.Suffix = `downed`
	} else if c.Aggro != nil && c.Aggro.UserId == viewingUserId {
		f.Suffix = `aggro`
	}

	return f
}

func (c *Character) SetAdjective(adj string, addToList bool) {
	if c.Adjectives == nil {
		c.Adjectives = []string{}
	}
	for i, a := range c.Adjectives {
		if a == adj {
			if addToList {
				return
			} else {
				c.Adjectives = append(c.Adjectives[:i], c.Adjectives[i+1:]...)
				return
			}
		}
	}
	if addToList {
		c.Adjectives = append(c.Adjectives, adj)
	}
}

func (c *Character) PruneCooldowns() {
	if len(c.Cooldowns) == 0 {
		return
	}

	c.Cooldowns.Prune()

}

func (c *Character) GetCooldown(trackingTag string) int {
	if c.Cooldowns == nil {
		c.Cooldowns = make(Cooldowns)
	}
	return c.Cooldowns[trackingTag]
}

func (c *Character) GetAllCooldowns() map[string]int {

	ret := map[string]int{}

	if c.Cooldowns == nil {
		return ret
	}

	for trackingTag, rounds := range c.Cooldowns {
		ret[trackingTag] = rounds
	}

	return ret
}

func (c *Character) TryCooldown(trackingTag string, cooldownRounds int) bool {
	if c.Cooldowns == nil {
		c.Cooldowns = make(Cooldowns)
	}

	return c.Cooldowns.Try(trackingTag, cooldownRounds)
}

func (c *Character) SetSetting(settingName string, settingValue string) {
	if c.Settings == nil {
		c.Settings = make(map[string]string)
	}

	if settingValue == "" {
		delete(c.Settings, settingName)
	} else {
		c.Settings[settingName] = settingValue
	}
}

func (c *Character) GetSetting(settingName string) string {
	if c.Settings == nil {
		c.Settings = make(map[string]string)
	}
	if settingValue, ok := c.Settings[settingName]; ok {
		return settingValue
	}
	return ""
}

func (c *Character) StoreItem(i items.Item) bool {
	if i.ItemId < 1 {
		return false
	}
	c.Items = append(c.Items, i)

	return true
}

func (c *Character) RemoveItem(i items.Item) bool {
	for j := len(c.Items) - 1; j >= 0; j-- {
		if c.Items[j].Equals(i) {
			c.Items = append(c.Items[:j], c.Items[j+1:]...)
			return true
		}
	}
	return false
}

func (c *Character) HandsRequired(i items.Item) int {

	if i.ItemId < 1 {
		return 0
	}

	iSpec := i.GetSpec()

	// Shooting weapnos don't benefit from creature size
	// when determining how many hands they require
	if iSpec.Subtype == items.Shooting {
		return iSpec.Hands
	}

	raceInfo := races.GetRace(c.RaceId)
	if raceInfo.Size == races.Large {
		return 1
	}

	if raceInfo.Size == races.Small {
		return iSpec.Hands + 1
	}

	return iSpec.Hands
}

// Copies over an existing item with a new item
// Returns true if successfully replaces an item
func (c *Character) UpdateItem(originalItm items.Item, replacement items.Item) bool {
	for j := len(c.Items) - 1; j >= 0; j-- {
		if c.Items[j].Equals(originalItm) {
			// If the number of uses remaining has decremented from the original item
			// The item gets destroyed from existence
			if originalItm.Uses >= 1 && replacement.Uses < 1 {
				c.Items = append(c.Items[:j], c.Items[j+1:]...)
			} else {
				c.Items[j] = replacement
			}
			return true
		}
	}
	return false
}

func (c *Character) UseItem(i items.Item) int {
	for j := len(c.Items) - 1; j >= 0; j-- {
		if c.Items[j].Equals(i) {
			usesLeft := c.Items[j].Uses
			if usesLeft > 0 {
				usesLeft--
			}
			if usesLeft <= 0 {
				c.Items = append(c.Items[:j], c.Items[j+1:]...)
			} else {
				c.Items[j].Uses = usesLeft
				c.Items[j].LastUsedRound = util.GetRoundCount()
			}

			return usesLeft
		}
	}

	return 0
}

func (c *Character) FindInBackpack(itemName string) (items.Item, bool) {

	if itemName == `` {
		return items.Item{}, false
	}

	closeMatchItem, matchItem := items.FindMatchIn(itemName, c.Items...)

	if matchItem.ItemId != 0 {
		return matchItem, true
	}

	if closeMatchItem.ItemId != 0 {
		return closeMatchItem, true
	}

	return items.Item{}, false
}

func (c *Character) FindOnBody(itemName string) (items.Item, bool) {

	if itemName == `` {
		return items.Item{}, false
	}

	partialMatch, fullMatch := items.FindMatchIn(itemName,
		c.Equipment.Weapon,
		c.Equipment.Offhand,
		c.Equipment.Head,
		c.Equipment.Neck,
		c.Equipment.Body,
		c.Equipment.Belt,
		c.Equipment.Gloves,
		c.Equipment.Ring,
		c.Equipment.Legs,
		c.Equipment.Feet)

	if fullMatch.ItemId != 0 {
		return fullMatch, true
	}

	if partialMatch.ItemId != 0 {
		return partialMatch, true
	}

	return items.Item{}, false
}

func (c *Character) GetSkills() map[string]int {
	skillResults := make(map[string]int)
	for skillName, skillLevel := range c.Skills {
		skillResults[skillName] = skillLevel
	}
	return skillResults
}

func (c *Character) SetSkill(skillName string, level int) {
	if c.Skills == nil {
		c.Skills = make(map[string]int)
	}
	skillName = strings.ToLower(skillName)

	if level == 0 {
		delete(c.Skills, skillName)
		return
	}

	c.Skills[skillName] = level
}

// Increases the skill training counter and returns the new value
func (c *Character) TrainSkill(skillName string, targetLevel ...int) int {
	if c.Skills == nil {
		c.Skills = make(map[string]int)
	}

	skillName = strings.ToLower(skillName)

	skillLevel := 0

	if lvl, ok := c.Skills[skillName]; ok {
		skillLevel = lvl
	}

	if len(targetLevel) > 0 {

		if skillLevel < targetLevel[0] {
			skillLevel = targetLevel[0]
		}

	} else if skillLevel < 4 {

		skillLevel++

	}

	c.Skills[skillName] = skillLevel

	return skillLevel
}

// Gets the current value of the skillname provided
func (c *Character) GetSkillLevel(skillName skills.SkillTag) int {
	if c.Skills == nil {
		c.Skills = make(map[string]int)
	}

	if level, ok := c.Skills[string(skillName)]; ok {
		return level
	}
	return 0
}

func (c *Character) GetSkillLevelCost(currentLevel int) int {
	return currentLevel
}

func (c *Character) GetTameCreatureSkill(userId int, creatureName string) int {

	skillValue := c.GetMiscData(`tameskill-` + creatureName)
	if sVal, ok := skillValue.(int); ok {
		return sVal
	}
	return -1

}

func (c *Character) GetMaxCharmedCreatures() int {
	lvl := c.GetSkillLevel(skills.Tame)
	return lvl + 1
}

func (c *Character) SetTameCreatureSkill(userId int, creatureName string, proficiency int) error {
	c.SetMiscData(`tameskill-`+creatureName, proficiency)
	return nil
}

func (c *Character) GetMemoryCapacity() int {
	return c.GetSkillLevel(skills.Map)*c.Stats.Smarts.Value + 5
}

func (c *Character) GetMapSprawlCapacity() int {
	return c.GetSkillLevel(skills.Map) + (c.Stats.Smarts.Value >> 2)
}

// Return all rooms the player can remember visiting
func (c *Character) GetRoomMemory() []int {
	mapHistory := c.GetMemoryCapacity()
	// return the last {mapHistory} items
	if len(c.roomHistory) > mapHistory {
		// return a copy of the last {mapHistory} items
		return append([]int{}, c.roomHistory[len(c.roomHistory)-mapHistory:]...)
	}
	// return a full copy
	return append([]int{}, c.roomHistory...)
}

// Return all rooms the player can remember visiting
func (c *Character) SetRoomMemory(newMem []int) {
	c.roomHistory = newMem
}

// Remember visiting a room. This may cause to forget an older room if the memory is full.
func (c *Character) RememberRoom(roomId int) {
	mapHistory := c.GetMemoryCapacity()
	if len(c.roomHistory) >= mapHistory*2 {
		// Prune out everything except {mapHistory}-1 items at the end
		c.roomHistory = c.roomHistory[len(c.roomHistory)-(mapHistory-1):]
	}
	c.roomHistory = append(c.roomHistory, roomId)
}

func (c *Character) IsQuestDone(questToken string) bool {
	testQuestId, _ := quests.TokenToParts(questToken)
	if c.QuestProgress == nil {
		c.QuestProgress = make(map[int]string)
	}

	stage := c.QuestProgress[testQuestId]

	return stage == `end`
}

func (c *Character) HasQuest(questToken string) bool {

	if c.QuestProgress == nil {
		c.QuestProgress = make(map[int]string)
	}

	testQuestId, testQuestStep := quests.TokenToParts(questToken)

	currentStep, ok := c.QuestProgress[testQuestId]
	if !ok {
		return false
	}

	// If on that step currently, then true
	if currentStep == testQuestStep {
		return true
	}

	currentToken := quests.PartsToToken(testQuestId, currentStep)

	// If the current token comes after the test token then they've already done that quest
	return quests.IsTokenAfter(questToken, currentToken)
}

func (c *Character) GetQuestProgress() map[int]string {

	if c.QuestProgress == nil {
		c.QuestProgress = make(map[int]string)
	}

	retMap := make(map[int]string)
	for questId, stepName := range c.QuestProgress {
		retMap[questId] = stepName
	}
	return retMap
}

func (c *Character) GiveQuestToken(questToken string) bool {

	if c.QuestProgress == nil {
		c.QuestProgress = make(map[int]string)
	}

	questId, newStep := quests.TokenToParts(questToken)
	currentProgress := c.QuestProgress[questId]

	currentToken := quests.PartsToToken(questId, currentProgress)

	if quests.IsTokenAfter(currentToken, questToken) {
		c.QuestProgress[questId] = newStep
		return true
	}

	return false
}

func (c *Character) ClearQuestToken(questToken string) {

	if c.QuestProgress == nil {
		c.QuestProgress = make(map[int]string)
	}

	questId, _ := quests.TokenToParts(questToken)

	delete(c.QuestProgress, questId)
}

func (c *Character) SetAggroRemote(exitName string, userId int, mobInstanceId int, aggroType AggroType, roundsWaitTime ...int) {
	c.SetAggro(userId, mobInstanceId, aggroType, roundsWaitTime...)
	c.Aggro.ExitName = exitName
}

func (c *Character) SetAggro(userId int, mobInstanceId int, aggroType AggroType, roundsWaitTime ...int) {

	var combatAddlWaitRounds int = 0

	if len(roundsWaitTime) > 0 {
		for _, waitAmt := range roundsWaitTime {
			combatAddlWaitRounds += waitAmt
		}
	} else {
		combatAddlWaitRounds = c.Equipment.Weapon.GetSpec().WaitRounds + c.Equipment.Offhand.GetSpec().WaitRounds
	}

	if aggroType == DefaultAttack {
		if c.Equipment.Weapon.GetSpec().Subtype == items.Shooting {
			aggroType = Shooting
		}
	}

	c.Aggro = &Aggro{
		UserId:        userId,
		MobInstanceId: mobInstanceId,
		Type:          aggroType,
		RoundsWaiting: combatAddlWaitRounds,
	}

}

func (c *Character) SetCast(roundsWaitTime int, sInfo SpellAggroInfo) {

	c.Aggro = &Aggro{
		Type:          SpellCast,
		RoundsWaiting: roundsWaitTime,
		SpellInfo:     sInfo,
	}

}

func (c *Character) EndAggro() {
	c.Aggro = nil
}

func (c *Character) IsAggro(targetUserId int, targetMobInstanceId int) bool {

	if c.Aggro != nil {

		if c.Aggro.MobInstanceId > 0 && c.Aggro.MobInstanceId == targetMobInstanceId {
			return true
		}

		if c.Aggro.UserId > 0 && c.Aggro.UserId == targetUserId {
			return true
		}

		if c.Aggro.Type == SpellCast {
			if len(c.Aggro.SpellInfo.TargetUserIds) > 0 {
				for _, uId := range c.Aggro.SpellInfo.TargetUserIds {
					if uId == targetUserId {
						return true
					}
				}
			}

			if len(c.Aggro.SpellInfo.TargetMobInstanceIds) > 0 {
				for _, mId := range c.Aggro.SpellInfo.TargetMobInstanceIds {
					if mId == targetMobInstanceId {
						return true
					}
				}
			}
		}

	}
	return false
}

func (c *Character) IsDisabled() bool {
	return c.Health <= 0
}

func (c *Character) HasBuffFlag(buffFlag buffs.Flag) bool {
	return c.Buffs.HasFlag(buffFlag, false)
}

func (c *Character) CancelBuffsWithFlag(buffFlag buffs.Flag) bool {
	if c.Buffs.HasFlag(buffFlag, true) {
		c.Validate()
		return true
	}
	return false
}

func (c *Character) HasBuff(buffId int) bool {
	return c.Buffs.HasBuff(buffId)
}

func (c *Character) AddBuff(buffId int, fromItem ...bool) error {
	buffId = int(math.Abs(float64(buffId)))
	if !c.Buffs.AddBuff(buffId, fromItem...) {
		return fmt.Errorf(`failed to add buff. target: "%s" buffId: %d`, c.Name, buffId)
	}
	c.Validate()
	return nil
}

func (c *Character) TrackBuffStarted(buffId int) {
	c.Buffs.Started(buffId)
}

func (c *Character) GetBuffs(buffId ...int) []*buffs.Buff {
	return c.Buffs.GetBuffs(buffId...)
}

func (c *Character) RemoveBuff(buffId int) {
	buffId = int(math.Abs(float64(buffId)))
	c.Buffs.RemoveBuff(buffId)
	c.Validate()
}

func (c *Character) ApplyHealthChange(healthChange int) int {
	oldHealth := c.Health
	newHealth := c.Health + healthChange
	if newHealth < 0 {
		c.CancelBuffsWithFlag(buffs.CancelIfCombat)
		if newHealth < -10 {
			newHealth = -10
		}
	} else if newHealth > c.HealthMax.Value {
		newHealth = c.HealthMax.Value
	}

	c.Health = newHealth
	return newHealth - oldHealth
}

func (c *Character) ApplyManaChange(manaChange int) {
	c.Mana += manaChange
}

func (c *Character) BarterPrice(startPrice int) int {
	factor := (float64(c.Stats.Perception.Value) / 3) / 100 // 100 = 33% discount, 0 = 0% discount, 300 = 100% discount
	if factor > .75 {
		factor = .75
	}
	return int(factor * float64(startPrice))
}

func (c *Character) XPTNL() int {
	return c.XPTL(c.Level)
}

// Amt TNL for a specific level
func (c *Character) XPTL(lvl int) int {
	fLvl := float64(lvl)
	return int(float32(1000+(fLvl*(fLvl*.75)*1000)) * c.TNLScale * float32(configs.GetConfig().XPScale))
}

// Returns the actual xp in regards to the current level/next level
func (c *Character) XPTNLActual() (currentXP int, tnlXP int) {
	currentLevelXP := c.XPTL(c.Level - 1)
	if c.Level == 1 {
		currentLevelXP = 0
	}
	nextLevelXP := c.XPTL(c.Level)
	tnlXP = nextLevelXP - currentLevelXP
	currentXP = c.Experience - currentLevelXP
	return currentXP, tnlXP
}

func (c *Character) LevelUp() (bool, stats.Statistics) {
	if c.XPTNL() > c.Experience {
		return false, stats.Statistics{}
	}

	var statsBefore stats.Statistics = c.Stats

	c.Level++
	c.TrainingPoints++
	c.StatPoints++

	c.Validate()

	var statsDelta stats.Statistics = c.Stats

	statsDelta.Strength.Value -= statsBefore.Strength.Value
	statsDelta.Speed.Value -= statsBefore.Speed.Value
	statsDelta.Smarts.Value -= statsBefore.Smarts.Value
	statsDelta.Vitality.Value -= statsBefore.Vitality.Value
	statsDelta.Mysticism.Value -= statsBefore.Mysticism.Value
	statsDelta.Perception.Value -= statsBefore.Perception.Value

	c.Health = c.HealthMax.Value
	c.Mana = c.ManaMax.Value

	return true, statsDelta
}

func (c *Character) Heal(hp int, mana int) {
	c.Health += hp
	if c.Health > c.HealthMax.Value {
		c.Health = c.HealthMax.Value
	}
	c.Mana += hp
	if c.Mana > c.ManaMax.Value {
		c.Mana = c.ManaMax.Value
	}
}

func (c *Character) HealthPerRound() int {
	healAmt := math.Round(float64(c.Stats.Vitality.Value)/8) +
		math.Round(float64(c.Level)/12) +
		1.0

	return int(healAmt)
}

func (c *Character) ManaPerRound() int {
	healAmt := math.Round(float64(c.Stats.Mysticism.Value)/8) +
		math.Round(float64(c.Level)/12) +
		1.0

	return int(healAmt)
}

// Where 1000 = a full round
func (c *Character) MovementCost() int {
	modifier := 3                             // by default they should be able to move 3 times per round.
	modifier += int(c.Level / 15)             // Every 15 levels, get an extra movement.
	modifier += int(c.Stats.Speed.Value / 15) // Every 15 speed, get an extra movement
	return int(1000 / modifier)
}

func (c *Character) RecalculateStats() {

	// Make sure racial base stats are set

	if raceInfo := races.GetRace(c.RaceId); raceInfo != nil {
		c.TNLScale = raceInfo.TNLScale
		c.Stats.Strength.Base = raceInfo.Stats.Strength.Base
		c.Stats.Speed.Base = raceInfo.Stats.Speed.Base
		c.Stats.Smarts.Base = raceInfo.Stats.Smarts.Base
		c.Stats.Vitality.Base = raceInfo.Stats.Vitality.Base
		c.Stats.Mysticism.Base = raceInfo.Stats.Mysticism.Base
		c.Stats.Perception.Base = raceInfo.Stats.Perception.Base
	}

	// Add any mods for equipment
	c.Stats.Strength.Mods = c.Equipment.StatMod("strength") + c.Buffs.StatMod("strength")
	c.Stats.Speed.Mods = c.Equipment.StatMod("speed") + c.Buffs.StatMod("speed")
	c.Stats.Smarts.Mods = c.Equipment.StatMod("smarts") + c.Buffs.StatMod("smarts")
	c.Stats.Vitality.Mods = c.Equipment.StatMod("vitality") + c.Buffs.StatMod("vitality")
	c.Stats.Mysticism.Mods = c.Equipment.StatMod("mysticism") + c.Buffs.StatMod("mysticism")
	c.Stats.Perception.Mods = c.Equipment.StatMod("perception") + c.Buffs.StatMod("perception")

	// Recalculate stats
	// Stats are basically:
	// level*base + training + mods
	c.Stats.Strength.Recalculate(c.Level)
	c.Stats.Speed.Recalculate(c.Level)
	c.Stats.Smarts.Recalculate(c.Level)
	c.Stats.Vitality.Recalculate(c.Level)
	c.Stats.Mysticism.Recalculate(c.Level)
	c.Stats.Perception.Recalculate(c.Level)

	// Set HP/MP maxes
	// This relies on the above stats so has to be calculated afterwards
	c.HealthMax.Mods = 5 +
		c.Buffs.StatMod("healthmax") + // Any sort of spell buffs etc. are just direct modifiers
		c.Equipment.StatMod("healthmax") + // However many points you have from equipment, you get 1 hp per point
		c.Level + // For every level you get 1 hp
		c.Stats.Vitality.Value*4 // for every vitality you get 3hp

	c.ManaMax.Mods = 4 +
		c.Buffs.StatMod("manamax") + // Any sort of spell buffs etc. are just direct modifiers
		c.Equipment.StatMod("manamax") + // However many points you have from equipment, you get 1 hp per point
		c.Level + // For every level you get 1 mp
		c.Stats.Mysticism.Value*3 // for every Mysticism you get 2mp

	// Set max action points
	c.ActionPointsMax.Mods = 200 // hard coded for now

	// Recalculate HP/MP stats
	c.HealthMax.Recalculate(c.Level)
	c.ManaMax.Recalculate(c.Level)
	c.ActionPointsMax.Recalculate(c.Level)

	// HP can't max less than 1, MP can't max less than 0
	if c.ManaMax.Value < 0 {
		c.ManaMax.Value = 0
	}
	if c.HealthMax.Value < 1 {
		c.HealthMax.Value = 1
	}
	if c.ActionPointsMax.Value < 50 {
		c.ActionPointsMax.Value = 50
	}
}

// AutoTrain() spends any training points for this character
func (c *Character) AutoTrain() {

	if c.StatPoints < 0 {
		return
	}

	for c.StatPoints > 0 {

		switch util.Rand(6) {
		case 0:
			c.Stats.Strength.Training++
		case 1:
			c.Stats.Speed.Training++
		case 2:
			c.Stats.Smarts.Training++
		case 3:
			c.Stats.Vitality.Training++
		case 4:
			c.Stats.Mysticism.Training++
		case 5:
			c.Stats.Perception.Training++
		}

		c.StatPoints--
	}

	c.Validate()

}

func (c *Character) CanDualWield() bool {

	if c.GetSkillLevel(skills.DualWield) > 0 {
		return true
	}
	return false
}

// Returns whether a correction was in order
func (c *Character) Validate(recalculateItemBuffs ...bool) error {

	if len(c.Description) == 0 {
		c.Description = c.Name + " seems thoroughly uninteresting."
	}

	if c.SpellBook == nil {
		c.SpellBook = make(map[string]int)
	}

	if c.Zone == "" {
		c.Zone = startingZone
	}

	if c.Name == "" {
		c.Name = defaultName
	}
	if c.Level < 1 {
		c.Level = 1
	}
	if c.Experience < 1 {
		c.Experience = 1
	}

	c.Buffs.Validate()

	// Do a stats recalc based on equipment, race, level, etc.
	c.RecalculateStats()

	// Recalculate health and mana

	if c.Mana > c.ManaMax.Value {
		c.Mana = c.ManaMax.Value
	}
	if c.Health > c.HealthMax.Value {
		c.Health = c.HealthMax.Value
	}

	if c.Health < -10 {
		c.Health = -10
	}

	if c.Mana < 0 {
		c.Mana = 0
	}

	c.Cooldowns.Prune()

	if c.Alignment < AlignmentMinimum {
		c.Alignment = AlignmentMinimum
	}

	if c.Alignment > AlignmentMaximum {
		c.Alignment = AlignmentMaximum
	}

	if raceInfo := races.GetRace(c.RaceId); raceInfo != nil {

		c.Equipment.EnableAll()

		// Are there slots that SHOULD be disabled?
		if len(raceInfo.DisabledSlots) > 0 {

			for _, disabledSlot := range raceInfo.DisabledSlots {

				var itemFoundInDisabledSlot items.Item = items.ItemDisabledSlot

				switch items.ItemType(disabledSlot) {
				case items.Weapon:
					if c.Equipment.Weapon.ItemId > 0 { // Did we find somethign in a disabled slot?
						itemFoundInDisabledSlot = c.Equipment.Weapon
					}
					c.Equipment.Weapon = items.ItemDisabledSlot
				case items.Offhand, items.Holdable:
					if c.Equipment.Offhand.ItemId > 0 { // Did we find somethign in a disabled slot?
						itemFoundInDisabledSlot = c.Equipment.Offhand
					}
					c.Equipment.Offhand = items.ItemDisabledSlot
				case items.Head:
					if c.Equipment.Head.ItemId > 0 { // Did we find somethign in a disabled slot?
						itemFoundInDisabledSlot = c.Equipment.Head
					}
					c.Equipment.Head = items.ItemDisabledSlot
				case items.Neck:
					if c.Equipment.Neck.ItemId > 0 { // Did we find somethign in a disabled slot?
						itemFoundInDisabledSlot = c.Equipment.Neck
					}
					c.Equipment.Neck = items.ItemDisabledSlot
				case items.Body:
					if c.Equipment.Body.ItemId > 0 { // Did we find somethign in a disabled slot?
						itemFoundInDisabledSlot = c.Equipment.Body
					}
					c.Equipment.Body = items.ItemDisabledSlot
				case items.Belt:
					if c.Equipment.Belt.ItemId > 0 { // Did we find somethign in a disabled slot?
						itemFoundInDisabledSlot = c.Equipment.Belt
					}
					c.Equipment.Belt = items.ItemDisabledSlot
				case items.Gloves:
					if c.Equipment.Gloves.ItemId > 0 { // Did we find somethign in a disabled slot?
						itemFoundInDisabledSlot = c.Equipment.Gloves
					}
					c.Equipment.Gloves = items.ItemDisabledSlot
				case items.Ring:
					if c.Equipment.Ring.ItemId > 0 { // Did we find somethign in a disabled slot?
						itemFoundInDisabledSlot = c.Equipment.Ring
					}
					c.Equipment.Ring = items.ItemDisabledSlot
				case items.Legs:
					if c.Equipment.Legs.ItemId > 0 { // Did we find somethign in a disabled slot?
						itemFoundInDisabledSlot = c.Equipment.Legs
					}
					c.Equipment.Legs = items.ItemDisabledSlot
				case items.Feet:
					if c.Equipment.Feet.ItemId > 0 { // Did we find somethign in a disabled slot?
						itemFoundInDisabledSlot = c.Equipment.Feet
					}
					c.Equipment.Feet = items.ItemDisabledSlot
				}

				if !itemFoundInDisabledSlot.IsDisabled() {
					c.StoreItem(itemFoundInDisabledSlot)
					slog.Debug("Disabled Check", "error", "Item found in disabled slot", "name", itemFoundInDisabledSlot.Name(), "slot", disabledSlot, "character", c.Name)
				}
			}

		}

	}

	if len(recalculateItemBuffs) > 0 && recalculateItemBuffs[0] {
		c.reapplyWornItemBuffs()
	}

	return nil
}

func (c *Character) Race() string {
	if r := races.GetRace(c.RaceId); r != nil {
		return r.Name
	}
	return `Ghostly Spirit`
}

func (c *Character) AlignmentName() string {

	if c.Alignment < AlignmentNeutral {
		// -80 to -100
		if c.Alignment <= AlignmentNeutral-80 {
			return `unholy`
		}
		// -60 to -79
		if c.Alignment <= AlignmentNeutral-60 {
			return `evil`
		}
		// -40 to -59
		if c.Alignment <= AlignmentNeutral-40 {
			return `corrupt`
		}
		// -20 to -39
		if c.Alignment <= AlignmentNeutral-20 {
			return `misguided`
		}

	} else if c.Alignment > AlignmentNeutral {

		// 80-100
		if c.Alignment >= AlignmentNeutral+80 {
			return `holy`
		}
		// 60 to 79
		if c.Alignment >= AlignmentNeutral+60 {
			return `good`
		}
		// 40 to 59
		if c.Alignment >= AlignmentNeutral+40 {
			return `virtuous`
		}
		// 20 to 39
		if c.Alignment >= AlignmentNeutral+40 {
			return `lawful`
		}

	}

	return `neutral`

}

func (c *Character) GetAllBackpackItems() []items.Item {
	return append([]items.Item{}, c.Items...)
}

func (c *Character) GetAllWornItems() []items.Item {
	wornItems := []items.Item{}
	if c.Equipment.Weapon.ItemId > 0 {
		wornItems = append(wornItems, c.Equipment.Weapon)
	}
	if c.Equipment.Offhand.ItemId > 0 {
		wornItems = append(wornItems, c.Equipment.Offhand)
	}
	if c.Equipment.Head.ItemId > 0 {
		wornItems = append(wornItems, c.Equipment.Head)
	}
	if c.Equipment.Neck.ItemId > 0 {
		wornItems = append(wornItems, c.Equipment.Neck)
	}
	if c.Equipment.Body.ItemId > 0 {
		wornItems = append(wornItems, c.Equipment.Body)
	}
	if c.Equipment.Belt.ItemId > 0 {
		wornItems = append(wornItems, c.Equipment.Belt)
	}
	if c.Equipment.Gloves.ItemId > 0 {
		wornItems = append(wornItems, c.Equipment.Gloves)
	}
	if c.Equipment.Ring.ItemId > 0 {
		wornItems = append(wornItems, c.Equipment.Ring)
	}
	if c.Equipment.Legs.ItemId > 0 {
		wornItems = append(wornItems, c.Equipment.Legs)
	}
	if c.Equipment.Feet.ItemId > 0 {
		wornItems = append(wornItems, c.Equipment.Feet)
	}
	return wornItems
}

func (c *Character) Wear(i items.Item) (returnItems []items.Item, newItemWorn bool) {

	spec := i.GetSpec()

	if spec.Type != items.Weapon && spec.Subtype != items.Wearable {
		return returnItems, false
	}

	iHandsRequired := c.HandsRequired(i)
	if iHandsRequired > 2 {
		return returnItems, false
	}

	// are botht he currently equipped weapon and this weapon claws?
	bothMartial := false
	if spec.Subtype == items.Claws && c.Equipment.Weapon.GetSpec().Subtype == items.Claws {
		bothMartial = true
	}

	canDualWield := c.CanDualWield()

	// Weapons can go in either hand.
	// Only do this if this is a 1 handed weapon
	if spec.Type == items.Weapon && iHandsRequired < 2 {

		// If they can dual wield
		if canDualWield || bothMartial {

			// If they have a weapon equippment and it is 1 handed
			if c.Equipment.Weapon.ItemId != 0 && c.HandsRequired(c.Equipment.Weapon) == 1 {
				// If nothing is in their offhand
				if c.Equipment.Offhand.ItemId == 0 {
					// Put it in the offhand.
					returnItems = append(returnItems, c.Equipment.Offhand)
					c.Equipment.Offhand = i

					c.reapplyWornItemBuffs()

					return returnItems, true
				}
			}

		}

	}

	// First handle weapon/offhand, since they are special cases
	switch spec.Type {
	case items.Weapon:
		if c.Equipment.Weapon.IsDisabled() { // Don't allow equipping on a disabled slot
			return returnItems, false
		}

		if !c.Equipment.Offhand.IsDisabled() { // Don't allow equipping on a disabled slot
			// If it's a 2 handed weapon, remove whatever is in the offhand
			if iHandsRequired == 2 || !canDualWield && c.Equipment.Offhand.GetSpec().Type == items.Weapon {
				returnItems = append(returnItems, c.Equipment.Offhand)
				c.Equipment.Offhand = items.Item{}
			}
		}

		returnItems = append(returnItems, c.Equipment.Weapon)
		c.Equipment.Weapon = i
	case items.Offhand, items.Holdable:
		if c.Equipment.Offhand.IsDisabled() { // Don't allow equipping on a disabled slot
			return returnItems, false
		}

		if !c.Equipment.Weapon.IsDisabled() { // Don't allow equipping on a disabled slot
			// If they have a 2h weapon equipped, remove it
			if c.HandsRequired(c.Equipment.Weapon) == 2 {
				returnItems = append(returnItems, c.Equipment.Weapon)
				c.Equipment.Weapon = items.Item{}
			}
		}
		returnItems = append(returnItems, c.Equipment.Offhand)
		c.Equipment.Offhand = i
	case items.Head:
		if c.Equipment.Head.IsDisabled() { // Don't allow equipping on a disabled slot
			return returnItems, false
		}
		returnItems = append(returnItems, c.Equipment.Head)
		c.Equipment.Head = i
	case items.Neck:
		if c.Equipment.Neck.IsDisabled() { // Don't allow equipping on a disabled slot
			return returnItems, false
		}
		returnItems = append(returnItems, c.Equipment.Neck)
		c.Equipment.Neck = i
	case items.Body:
		if c.Equipment.Body.IsDisabled() { // Don't allow equipping on a disabled slot
			return returnItems, false
		}
		returnItems = append(returnItems, c.Equipment.Body)
		c.Equipment.Body = i
	case items.Belt:
		if c.Equipment.Belt.IsDisabled() { // Don't allow equipping on a disabled slot
			return returnItems, false
		}
		returnItems = append(returnItems, c.Equipment.Belt)
		c.Equipment.Belt = i
	case items.Gloves:
		if c.Equipment.Gloves.IsDisabled() { // Don't allow equipping on a disabled slot
			return returnItems, false
		}
		returnItems = append(returnItems, c.Equipment.Gloves)
		c.Equipment.Gloves = i
	case items.Ring:
		if c.Equipment.Ring.IsDisabled() { // Don't allow equipping on a disabled slot
			return returnItems, false
		}
		returnItems = append(returnItems, c.Equipment.Ring)
		c.Equipment.Ring = i
	case items.Legs:
		if c.Equipment.Legs.IsDisabled() { // Don't allow equipping on a disabled slot
			return returnItems, false
		}
		returnItems = append(returnItems, c.Equipment.Legs)
		c.Equipment.Legs = i
	case items.Feet:
		if c.Equipment.Feet.IsDisabled() { // Don't allow equipping on a disabled slot
			return returnItems, false
		}
		returnItems = append(returnItems, c.Equipment.Feet)
		c.Equipment.Feet = i
	default:
		return returnItems, false
	}

	c.reapplyWornItemBuffs(returnItems...)

	return returnItems, true
}

func (c *Character) RemoveFromBody(i items.Item) bool {

	if i.Equals(c.Equipment.Weapon) {
		c.Equipment.Weapon = items.Item{}
	} else if i.Equals(c.Equipment.Offhand) {
		c.Equipment.Offhand = items.Item{}
	} else if i.Equals(c.Equipment.Head) {
		c.Equipment.Head = items.Item{}
	} else if i.Equals(c.Equipment.Neck) {
		c.Equipment.Neck = items.Item{}
	} else if i.Equals(c.Equipment.Body) {
		c.Equipment.Body = items.Item{}
	} else if i.Equals(c.Equipment.Belt) {
		c.Equipment.Belt = items.Item{}
	} else if i.Equals(c.Equipment.Gloves) {
		c.Equipment.Gloves = items.Item{}
	} else if i.Equals(c.Equipment.Ring) {
		c.Equipment.Ring = items.Item{}
	} else if i.Equals(c.Equipment.Legs) {
		c.Equipment.Legs = items.Item{}
	} else if i.Equals(c.Equipment.Feet) {
		c.Equipment.Feet = items.Item{}
	} else {
		return false
	}

	c.reapplyWornItemBuffs(i)

	return true
}

func (c *Character) reapplyWornItemBuffs(removedItems ...items.Item) {

	buffIdCount := map[int]int{}

	// Track any buffs that come from an item
	// If these don't show up as still being required by an item (such as a yaml file was changed)
	// This will cause them to be removed.
	for _, b := range c.Buffs.List {
		if b.ItemBuff {
			buffIdCount[b.BuffId] = 0
		}
	}

	// Make a list of all item buffs provided by existing worn items
	for _, itm := range c.GetAllWornItems() {
		spec := itm.GetSpec()
		for _, buffId := range spec.WornBuffIds {
			buffIdCount[buffId] = buffIdCount[buffId] + 1
		}

	}
	// Remove any buffs that come specifically from item
	for _, removedItem := range removedItems {
		iSpec := removedItem.GetSpec()
		if len(iSpec.WornBuffIds) > 0 {
			for _, buffId := range iSpec.WornBuffIds {
				buffIdCount[buffId] = buffIdCount[buffId] - 1
			}
		}
	}

	for buffId, ct := range buffIdCount {
		if ct < 1 {
			c.RemoveBuff(buffId)
		} else {
			c.AddBuff(buffId, true)
		}
	}
}
