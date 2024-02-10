package mobs

import (
	"fmt"
	"log/slog"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/fileloader"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/races"
	"github.com/volte6/mud/util"
	"gopkg.in/yaml.v2"
)

var (
	mobs            map[int]*Mob = map[int]*Mob{}
	allMobNames                  = []string{}
	mobInstances    map[int]*Mob = map[int]*Mob{}
	instanceCounter int          = 0
	mobMutex        sync.RWMutex
	mobsHatePlayers map[string]map[int]int = map[string]map[int]int{}
)

const (
	mobDataFilesFolderPath = "_datafiles/mobs"
)

type ItemTrade struct {
	AcceptedItemIds []int         `yaml:"accepteditemids,omitempty,flow"` // Must provide every item id in this list.
	AcceptedGold    int           `yaml:"acceptedgold,omitempty,flow"`    // Must provide at least this much gold.
	PrizeItemIds    []int         `yaml:"prizeitemids,omitempty,flow"`    // Will give these items in exchange.
	PrizeBuffIds    []int         `yaml:"prizebuffids,omitempty,flow"`    // Will give these buffs in exchange.
	PrizeRoomId     int           `yaml:"prizeroomid,omitempty,flow"`     // Will move player to this room in exchange.
	PrizeQuestIds   []string      `yaml:"prizequestids,omitempty,flow"`   // What quest id's will be awarded?
	PrizeGold       int           `yaml:"prizegold,omitempty,flow"`       // How much gold are they given?
	PrizeCommands   []string      `yaml:"prizecommands,omitempty,flow"`   // What commands will be executed?
	GivenItems      map[int][]int `yaml:"-"`                              // key = userId, value = Items given. Should only contain items from AcceptedItemIds
	GivenGold       map[int]int   `yaml:"-"`                              // key = userId, value = how much gold is given
}

type AskMob struct {
	IfQuest       string
	IfNotQuest    string
	AskNouns      []string
	ReplyCommands []string
}

type MobForHire struct {
	MobId    MobId
	Price    int
	Quantity int
}
type MobId int // Creating a custom type to help prevent confusion over MobId and MobInstanceId

type Mob struct {
	MobId           MobId
	Zone            string      `yaml:"zone,omitempty"`
	ItemDropChance  int         // chance in 100
	ActivityLevel   int         `yaml:"activitylevel,omitempty"` // 1 - 10%, 10 = 100%
	InstanceId      int         `yaml:"-"`
	HomeRoomId      int         `yaml:"-"`
	Hostile         bool        // whether they attack on sight
	IsMerchant      bool        `yaml:"ismerchant,omitempty"` // Are they a merchant?
	LastIdleCommand uint8       `yaml:"-"`                    // Track what hte last used idlecommand was
	BoredomCounter  uint8       `yaml:"-"`                    // how many rounds have passed since this mob has seen a player
	Groups          []string    // What group do they identify with? Helps with teamwork
	Hates           []string    `yaml:"hates,omitempty"`        // What NPC groups or races do they hate and probably fight if encountered?
	IdleCommands    []string    `yaml:"idlecommands,omitempty"` // Commands they may do while idle (not in combat)
	AngryCommands   []string    // randomly chosen to queue when they are angry/entering combat.
	CombatCommands  []string    `yaml:"combatcommands,omitempty"` // Commands they may do while in combat
	DamageTaken     map[int]int `yaml:"-"`                        // key = who, value = how much
	Character       characters.Character
	ShopStock       map[int]int    `yaml:"shopstock,omitempty"`
	ShopServants    []MobForHire   `yaml:"shopservants,omitempty"`
	MaxWander       int            `yaml:"maxwander,omitempty"`  // Max rooms to wander from home
	GoingHome       bool           `yaml:"-"`                    // WHether they are trying to get home
	RoomStack       []int          `yaml:"-"`                    // Stack of rooms to get back home
	PreventIdle     bool           `yaml:"-"`                    // Whether they can't possibly be idle
	ItemTrades      []ItemTrade    `yaml:"itemtrades,omitempty"` // one or more sets of objects they will trade for other objects.
	AskSubjects     []AskMob       `yaml:"asksubjects,omitempty"`
	ScriptTag       string         `yaml:"scripttag"` // Script for this mob: mobs/frostfang/scripts/{mobId}-{ScriptTag}.js
	datastub        map[string]any // Generic storage stub for maintaining state between behaviors
}

// If a character is provided, will only return true if character has the quest
func (m *Mob) HasQuestWaiting(c *characters.Character) bool {

	for _, subject := range m.AskSubjects {

		if len(subject.IfQuest) > 0 {
			if c == nil || c.HasQuest(subject.IfQuest) {
				return true
			}
		} else if len(subject.IfNotQuest) > 0 {
			if c == nil || c.HasQuest(subject.IfNotQuest) {
				return !c.IsQuestDone(subject.IfNotQuest)
			}
		}
	}

	return false
}

func GetAllMobNames() []string {
	return append([]string{}, allMobNames...)
}

func MobIdByName(mobName string) MobId {

	mobMutex.Lock()
	defer mobMutex.Unlock()

	match, partial := util.FindMatchIn(mobName, allMobNames...)
	if match == "" {
		match = partial
	}
	if match == "" {
		return 0
	}

	for _, m := range mobs {
		if m.Character.Name == match {
			return m.MobId
		}
	}

	for _, m := range mobs {
		if strings.HasPrefix(m.Character.Name, match) {
			return m.MobId
		}
	}

	for _, m := range mobs {
		if strings.Contains(m.Character.Name, match) {
			return m.MobId
		}
	}

	return 0
}

func NewMobById(mobId MobId, homeRoomId int) *Mob {

	mobMutex.Lock()
	defer mobMutex.Unlock()

	if m, ok := mobs[int(mobId)]; ok {

		instanceCounter++

		mob := *m // Make a copy of the mob

		mob.HomeRoomId = homeRoomId
		mob.Character.RoomId = homeRoomId
		mob.InstanceId = instanceCounter
		mob.DamageTaken = make(map[int]int)
		mob.Character.StatPoints = mob.Character.Level
		mob.Character.Level--
		mob.Character.Experience = mob.Character.XPTNL()
		mob.Character.Level++
		// Apply training for those stats
		mob.Character.AutoTrain()
		mob.Character.Health = mob.Character.HealthMax.Value
		mob.Character.Mana = mob.Character.ManaMax.Value

		mob.Character.Buffs = buffs.New()

		for idx, _ := range mob.Character.Items {
			mob.Character.Items[idx].Validate()
		}

		mob.Character.Equipment.Weapon.Validate()
		mob.Character.Equipment.Offhand.Validate()
		mob.Character.Equipment.Head.Validate()
		mob.Character.Equipment.Neck.Validate()
		mob.Character.Equipment.Body.Validate()
		mob.Character.Equipment.Belt.Validate()
		mob.Character.Equipment.Gloves.Validate()
		mob.Character.Equipment.Ring.Validate()
		mob.Character.Equipment.Legs.Validate()
		mob.Character.Equipment.Feet.Validate()

		mob.Validate()

		// Save the mob instance
		mobInstances[mob.InstanceId] = &mob

		return mobInstances[mob.InstanceId]
	}
	return nil
}

func GetMobSpec(mobId MobId) *Mob {
	if m, ok := mobs[int(mobId)]; ok {
		mob := *m // Make a copy of the mob
		return &mob
	}
	return nil
}

func GetInstance(instanceId int) *Mob {

	mobMutex.RLock()
	defer mobMutex.RUnlock()

	if m, ok := mobInstances[instanceId]; ok {
		return m
	}
	return nil
}

func GetAllMobInstanceIds() []int {

	mobMutex.RLock()
	defer mobMutex.RUnlock()

	ids := make([]int, 0)
	for id := range mobInstances {
		ids = append(ids, id)
	}
	return ids
}

func DestroyInstance(instanceId int) {
	mobMutex.Lock()
	defer mobMutex.Unlock()

	delete(mobInstances, instanceId)
}

func (m *Mob) Despawns() bool {
	if m.IsMerchant || len(m.ItemTrades) > 0 || len(m.AskSubjects) > 0 {
		return false
	}
	return true
}

func (m *Mob) GetData(dataname string) (any, bool) {
	if m.datastub == nil {
		m.datastub = make(map[string]any)
	}
	if val, ok := m.datastub[dataname]; ok {
		return val, true
	}
	return nil, false
}

func (m *Mob) SetData(dataname string, data any) {
	if m.datastub == nil {
		m.datastub = make(map[string]any)
	}
	m.datastub[dataname] = data
}

func (r *Mob) GetSellPrice(item items.Item) int {

	if item.IsSpecial() {
		return 0
	}

	itemType := item.GetSpec().Type
	itemSubtype := item.GetSpec().Subtype
	value := item.GetSpec().Value
	likesType := false
	likesSubtype := false

	if len(r.IdleCommands) > 0 {
		for _, cmdStr := range r.IdleCommands {
			if strings.HasPrefix(cmdStr, "restock ") {
				cmdStr = strings.TrimPrefix(cmdStr, "restock ")
				cmdParts := strings.Split(cmdStr, "/")
				itemId, _ := strconv.Atoi(cmdParts[0])
				if iSpec := items.New(itemId); iSpec.ItemId != 0 {
					if iSpec.GetSpec().Type == itemType {
						likesType = true
					}
					if iSpec.GetSpec().Subtype == itemSubtype {
						likesSubtype = true
					}
				}
			}
			if likesType && likesSubtype {
				break
			}
		}
	}

	if len(r.Character.Items) > 0 {
		for _, item := range r.Character.Items {
			if item.GetSpec().Type == itemType {
				likesType = true
			}
			if item.GetSpec().Subtype == itemSubtype {
				likesSubtype = true
			}
			if likesType && likesSubtype {
				break
			}
		}
	}

	for _, item := range r.Character.Equipment.GetAllItems() {
		if item.GetSpec().Type == itemType {
			likesType = true
		}
		if item.GetSpec().Subtype == itemSubtype {
			likesSubtype = true
		}
		if likesType && likesSubtype {
			break
		}
	}

	priceScale := 0.0
	if likesType {
		priceScale += 0.1
	}
	if likesSubtype {
		priceScale += 0.1
	}
	return int(math.Ceil(float64(value) * priceScale))
}

func (r *Mob) HatesRace(raceName string) bool {
	raceName = strings.ToLower(raceName)
	for _, hateGroup := range r.Hates {
		if hateGroup == raceName {
			return true
		}
	}
	return false
}

func (r *Mob) HatesMob(m *Mob) bool {
	if r.MobId == m.MobId {
		return false // Can't hate exact same as self
	}

	mRace := races.GetRace(m.Character.RaceId)
	raceName := strings.ToLower(mRace.Name)
	for _, rGroup := range r.Groups {
		if rGroup == raceName {
			return true
		}
		for _, mGroup := range m.Groups {
			if rGroup == mGroup {
				return false // Can't hate groups its part of.
			}
		}
	}
	// Loop through groups it hates and if it finds a match, return true
	for _, groupName := range r.Hates {
		if groupName == `*` { // If * it hates all groups
			return true
		}
		for _, mGroup := range m.Groups {
			if groupName == mGroup {
				return true
			}
		}
	}
	return false
}

func (m *Mob) GetAngryCommand() string {

	// First check if the mob has a specific action
	if len(m.AngryCommands) > 0 {
		return m.AngryCommands[util.Rand(len(m.AngryCommands))]
	}

	// default to race based actions
	r := races.GetRace(m.Character.RaceId)
	actionCt := len(r.AngryCommands)
	if actionCt > 0 {
		return r.AngryCommands[util.Rand(actionCt)]
	}
	return ``
}

func (r *Mob) IsAlly(m *Mob) bool {

	if m.MobId == r.MobId {
		return true // Auto ally with own kind
	}

	if len(m.Groups) == 0 && len(r.Groups) == 0 {
		return true // No allegiance on either side, consider an ally for now
	}

	// If they both belong to factions/groups, check for matches
	if len(m.Groups) > 0 && len(r.Groups) > 0 {
		// Look for a group match
		for _, testGroup := range m.Groups {
			for _, targetGroup := range r.Groups {
				if testGroup == targetGroup {
					return true
				}
			}
		}
	}

	return false
}

func (r *Mob) Id() int {
	return int(r.MobId)
}

func (r *Mob) Validate() error {
	if r.ActivityLevel < 1 {
		r.ActivityLevel = 1
	}
	if r.ActivityLevel > 10 {
		r.ActivityLevel = 10
	}
	if r.IsMerchant {
		if r.ShopStock == nil {
			r.ShopStock = make(map[int]int)
		}
	}

	if len(r.ItemTrades) > 0 {
		for idx, trade := range r.ItemTrades {
			if trade.GivenItems == nil {
				trade.GivenItems = make(map[int][]int)
			}
			if trade.GivenGold == nil {
				trade.GivenGold = make(map[int]int)
			}
			r.ItemTrades[idx] = trade
		}
	}

	r.Character.Validate()
	return nil
}

func (m *Mob) Filename() string {
	return fmt.Sprintf("%d.yaml", m.Id())
}

func (m *Mob) Filepath() string {
	zone := ZoneNameSanitize(m.Zone)
	return util.FilePath(zone, `/`, fmt.Sprintf("%d.yaml", m.Id()))
}

func (r *Mob) Save() error {

	fileName := r.Filename()

	bytes, err := yaml.Marshal(r)
	if err != nil {
		return err
	}

	saveFilePath := util.FilePath(mobDataFilesFolderPath, `/`, fmt.Sprintf("%s.yaml", fileName))

	err = os.WriteFile(saveFilePath, bytes, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mob) GetScript() string {

	scriptPath := m.GetScriptPath()
	// Load the script into a string
	if _, err := os.Stat(scriptPath); err == nil {
		if bytes, err := os.ReadFile(scriptPath); err == nil {
			return string(bytes)
		}
	}

	return ``
}

func (m *Mob) GetScriptPath() string {
	// Load any script for the room

	mobFilePath := m.Filename()

	newExt := `.js`
	if m.ScriptTag != `` {
		newExt = fmt.Sprintf(`-%s.js`, m.ScriptTag)
	}

	scriptFilePath := `scripts/` + strings.Replace(mobFilePath, `.yaml`, newExt, 1)

	fullScriptPath := strings.Replace(mobDataFilesFolderPath+`/`+m.Filepath(),
		mobFilePath,
		scriptFilePath,
		1)

	return util.FilePath(fullScriptPath)
}

func ReduceHostility() {

	mobMutex.Lock()
	defer mobMutex.Unlock()

	for groupName, group := range mobsHatePlayers {
		for userId, rounds := range group {
			rounds--
			if rounds < 1 {
				delete(mobsHatePlayers[groupName], userId)
			} else {
				mobsHatePlayers[groupName][userId] = rounds
			}
		}
		if len(mobsHatePlayers[groupName]) < 1 {
			delete(mobsHatePlayers, groupName)
		}
	}
}

func IsHostile(groupName string, userId int) bool {

	mobMutex.RLock()
	defer mobMutex.RUnlock()

	if _, ok := mobsHatePlayers[groupName]; !ok {
		return false
	}

	if _, ok := mobsHatePlayers[groupName][userId]; !ok {
		return false
	}

	return true
}

func MakeHostile(groupName string, userId int, rounds int) {

	mobMutex.Lock()
	defer mobMutex.Unlock()

	if _, ok := mobsHatePlayers[groupName]; !ok {
		mobsHatePlayers[groupName] = make(map[int]int)
		mobsHatePlayers[groupName][userId] = rounds
		return
	}
	if mobsHatePlayers[groupName][userId] < rounds {
		mobsHatePlayers[groupName][userId] = rounds
	}
}

func ZoneNameSanitize(zone string) string {
	if zone == "" {
		return ""
	}
	// Convert spaces to underscores
	zone = strings.ReplaceAll(zone, " ", "_")
	// Lowercase it all, and add a slash at the end
	return strings.ToLower(zone)
}

// file self loads due to init()
func LoadDataFiles() {

	start := time.Now()

	var err error
	mobs, err = fileloader.LoadAllFlatFiles[int, *Mob](mobDataFilesFolderPath)
	if err != nil {
		panic(err)
	}

	for _, mob := range mobs {

		mob.Character.CacheDescription()

		allMobNames = append(allMobNames, mob.Character.Name)
	}

	slog.Info("mobs.LoadDataFiles()", "loadedCount", len(mobs), "Time Taken", time.Since(start))

}
