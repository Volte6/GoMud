package rooms

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/gametime"
	"github.com/volte6/gomud/internal/items"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/mutators"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

const roomDataFilesPath = "_datafiles/rooms"
const visitorTrackingTimeout = 180  // 180 seconds (3 minutes?)
const roomUnloadTimeoutRounds = 450 // 1800 seconds (30 minutes) / 4 seconds (1 round) = 450 rounds
const defaultMapSymbol = `•`

var (
	MapSymbolOverrides = map[string]string{
		"*": defaultMapSymbol,
		//"•": "*",
	}
)

type FindFlag uint16
type VisitorType string

const (
	AffectsNone   = ""
	AffectsPlayer = "player" // Does it affect only the player who triggered it?
	AffectsRoom   = "room"   // Does it affect everyone in the room?

	// Useful for finding mobs/players
	FindCharmed        FindFlag = 0b00000000001 // charmed
	FindNeutral        FindFlag = 0b00000000010 // Not aggro, not charmed, not Hostile
	FindFightingPlayer FindFlag = 0b00000000100 // aggro vs. a player
	FindFightingMob    FindFlag = 0b00000001000 // aggro vs. a mob
	FindHostile        FindFlag = 0b00000010000 // will auto-attack players
	FindMerchant       FindFlag = 0b00000100000 // is a merchant
	FindDowned         FindFlag = 0b00001000000 // hp < 1
	FindBuffed         FindFlag = 0b00010000000 // has a buff
	FindHasLight       FindFlag = 0b00100000000 // has a light source
	FindHasPet         FindFlag = 0b01000000000 // has a pet
	FindNative         FindFlag = 0b10000000000 // spawns in this room

	// Combinatorial flags
	FindFighting          = FindFightingPlayer | FindFightingMob // Currently in combat (aggro)
	FindIdle              = FindCharmed | FindNeutral            // Not aggro or hostile
	FindAll      FindFlag = 0b111111111
	// Visitor types
	VisitorUser = "user"
	VisitorMob  = "mob"
)

type Room struct {
	//mutex
	RoomId            int        // a unique numeric index of the room. Also the filename.
	Zone              string     // zone is a way to partition rooms into groups. Also into folders.
	ZoneConfig        ZoneConfig `yaml:"zoneconfig,omitempty"`      // If non-null is a root room.
	IsBank            bool       `yaml:"isbank,omitempty"`          // Is this a bank room? If so, players can deposit/withdraw gold here.
	IsStorage         bool       `yaml:"isstorage,omitempty"`       // Is this a storage room? If so, players can add/remove objects here.
	IsCharacterRoom   bool       `yaml:"ischaracterroom,omitempty"` // Is this a room where characters can create new characters to swap between them?
	Title             string
	Description       string
	MapSymbol         string               `yaml:"mapsymbol,omitempty"`  // The symbol to use when generating a map of the zone
	MapLegend         string               `yaml:"maplegend,omitempty"`  // The text to display in the legend for this room. Should be one word.
	Biome             string               `yaml:"biome,omitempty"`      // The biome of the room. Used for weather generation.
	Containers        map[string]Container `yaml:"containers,omitempty"` // If this room has a chest, what is in it?
	Exits             map[string]RoomExit
	ExitsTemp         map[string]TemporaryRoomExit   `yaml:"-"`               // Temporary exits that will be removed after a certain time. Don't bother saving on sever shutting down.
	Nouns             map[string]string              `yaml:"nouns,omitempty"` // Interesting nouns to highlight in the room or reveal on succesful searches.
	Items             []items.Item                   `yaml:"items,omitempty"`
	Stash             []items.Item                   `yaml:"stash,omitempty"`             // list of items in the room that are not visible to players
	Gold              int                            `yaml:"gold,omitempty"`              // How much gold is on the ground?
	SpawnInfo         []SpawnInfo                    `yaml:"spawninfo,omitempty"`         // key is creature ID, value is spawn chance
	SkillTraining     map[string]TrainingRange       `yaml:"skilltraining,omitempty"`     // list of skills that can be trained in this room
	Signs             []Sign                         `yaml:"sign,omitempty"`              // list of scribbles in the room
	IdleMessages      []string                       `yaml:"idlemessages,omitempty"`      // list of messages that can be displayed to players in the room
	LastIdleMessage   uint8                          `yaml:"-"`                           // index of the last idle message displayed
	LongTermDataStore map[string]any                 `yaml:"longtermdatastore,omitempty"` // Long term data store for the room
	Mutators          mutators.MutatorList           `yaml:"mutators,omitempty"`          // mutators this room spawns with.
	Effects           map[EffectType]AreaEffect      `yaml:"-"`
	players           []int                          `yaml:"-"` // list of user IDs currently in the room
	mobs              []int                          `yaml:"-"` // list of mob instance IDs currently in the room. Does not get saved.
	visitors          map[VisitorType]map[int]uint64 `yaml:"-"` // list of user IDs that have visited this room, and the last round they did
	lastVisited       uint64                         `yaml:"-"` // last round a visitor was in the room
	tempDataStore     map[string]any                 `yaml:"-"` // Temporary data store for the room
}

type TrainingRange struct {
	Min int
	Max int
}

func NewRoom(zone string) *Room {
	r := &Room{
		RoomId:        GetNextRoomId(),
		Zone:          zone,
		Title:         "An empty room.",
		Description:   "This is an empty room that was never given a description.",
		MapSymbol:     ``,
		Exits:         make(map[string]RoomExit),
		Effects:       map[EffectType]AreaEffect{},
		players:       []int{},
		visitors:      make(map[VisitorType]map[int]uint64),
		tempDataStore: make(map[string]any),
	}

	SetNextRoomId(r.RoomId + 1)

	return r
}

func (r *Room) SendTextCommunication(txt string, excludeUserIds ...int) {

	events.AddToQueue(events.Message{
		RoomId:          r.RoomId,
		Text:            txt + "\n",
		ExcludeUserIds:  excludeUserIds,
		IsQuiet:         false,
		IsCommunication: true,
	})

}

func (r *Room) SendText(txt string, excludeUserIds ...int) {

	events.AddToQueue(events.Message{
		RoomId:         r.RoomId,
		Text:           txt + "\n",
		ExcludeUserIds: excludeUserIds,
		IsQuiet:        false,
	})

}

func (r *Room) SendTextToExits(txt string, isQuiet bool, excludeUserIds ...int) {

	testExitIds := []int{}
	for _, rExit := range r.Exits {
		testExitIds = append(testExitIds, rExit.RoomId)
	}
	for _, tExit := range r.ExitsTemp {
		testExitIds = append(testExitIds, tExit.RoomId)
	}

	for _, roomId := range testExitIds {

		tgtRoom := LoadRoom(roomId)
		if tgtRoom == nil {
			continue
		}

		for exitName, tExit := range tgtRoom.Exits {
			if tExit.RoomId != r.RoomId {
				continue
			}

			events.AddToQueue(events.Message{
				RoomId:         tgtRoom.RoomId,
				Text:           fmt.Sprintf(`(From <ansi fg="exit">%s</ansi>) `, exitName) + txt + "\n",
				IsQuiet:        isQuiet,
				ExcludeUserIds: excludeUserIds,
			})
		}

	}

}

func (r *Room) IsBurning() bool {
	return r.HasEffect(Wildfire)
}

func (r *Room) SetLongTermData(key string, value any) {

	if r.LongTermDataStore == nil {
		r.LongTermDataStore = make(map[string]any)
	}

	if value == nil {
		delete(r.LongTermDataStore, key)
		return
	}
	r.LongTermDataStore[key] = value
}

func (r *Room) GetLongTermData(key string) any {

	if r.LongTermDataStore == nil {
		r.LongTermDataStore = make(map[string]any)
	}

	if value, ok := r.LongTermDataStore[key]; ok {
		return value
	}
	return nil
}

func (r *Room) SetTempData(key string, value any) {

	if r.tempDataStore == nil {
		r.tempDataStore = make(map[string]any)
	}

	if value == nil {
		delete(r.tempDataStore, key)
		return
	}
	r.tempDataStore[key] = value
}

func (r *Room) GetTempData(key string) any {

	if r.tempDataStore == nil {
		r.tempDataStore = make(map[string]any)
	}

	if value, ok := r.tempDataStore[key]; ok {
		return value
	}
	return nil
}

func (r *Room) GetScript() string {

	scriptPath := r.GetScriptPath()

	// Load the script into a string
	if _, err := os.Stat(scriptPath); err == nil {
		if bytes, err := os.ReadFile(scriptPath); err == nil {
			return string(bytes)
		}
	}

	return ``
}

func (r *Room) GetScriptPath() string {

	// Load any script for the room
	return strings.Replace(roomDataFilesPath+`/`+r.Filepath(), `.yaml`, `.js`, 1)
}

func (r *Room) FindTemporaryExitByUserId(userId int) (TemporaryRoomExit, bool) {

	if r.ExitsTemp != nil {
		for _, v := range r.ExitsTemp {
			if v.UserId == userId {
				return v, true
			}
		}
	}

	return TemporaryRoomExit{}, false
}

func (r *Room) RemoveTemporaryExit(t TemporaryRoomExit) bool {

	if r.ExitsTemp == nil {
		return false
	}

	for k, v := range r.ExitsTemp {
		if v.UserId == t.UserId && v.Title == t.Title && t.RoomId == v.RoomId {
			delete(r.ExitsTemp, k)
			return true
		}
	}

	return false
}

// Can't add twoof the same exitName
// Will return false if it already exists
func (r *Room) AddTemporaryExit(exitName string, t TemporaryRoomExit) bool {

	if r.ExitsTemp == nil {
		r.ExitsTemp = make(map[string]TemporaryRoomExit)
	}

	if len(t.Title) == 0 {
		t.Title = exitName
	}
	if _, ok := r.ExitsTemp[exitName]; ok {
		return false
	}
	r.ExitsTemp[exitName] = t
	return true
}

// applies buffs to any players in the room that don't
// already have it
func (r *Room) ApplyBuffIdToPlayers(buffId ...int) {

	if len(buffId) == 0 {
		return
	}

	for _, uid := range r.GetPlayers() {

		if u := users.GetByUserId(uid); u != nil {

			for _, bId := range buffId {
				if u.Character.HasBuff(bId) {
					continue
				}
				u.AddBuff(bId)
			}
		}

	}

}

// applies buffs to any mobs in the room that don't
// already have it
func (r *Room) ApplyBuffIdToMobs(buffId ...int) {

	if len(buffId) == 0 {
		return
	}

	for _, miid := range r.GetMobs() {

		if m := mobs.GetInstance(miid); m != nil {

			for _, bId := range buffId {
				if m.Character.HasBuff(bId) {
					continue
				}
				m.AddBuff(bId)
			}
		}

	}

}

// applies buffs to any mobs in the room that don't
// already have it
func (r *Room) ApplyBuffIdToNativeMobs(buffId ...int) {

	if len(buffId) == 0 {
		return
	}

	for _, miid := range r.GetMobs(FindNative) {

		if m := mobs.GetInstance(miid); m != nil {

			for _, bId := range buffId {
				if m.Character.HasBuff(bId) {
					continue
				}
				m.AddBuff(bId)
			}
		}

	}

}

func (r *Room) SpawnTempContainer(name string, duration string, lockDifficulty int, trapBuffIds ...int) string {

	c := Container{}

	gd := gametime.GetDate(util.GetRoundCount())
	c.DespawnRound = gd.AddPeriod(duration)

	c.Lock.Difficulty = uint8(lockDifficulty)

	if len(trapBuffIds) > 0 {
		c.Lock.TrapBuffIds = trapBuffIds
	}

	containerName := name

	// make sure name is unique
	i := 1
	_, ok := r.Containers[containerName]
	for ok {
		containerName = name + `-` + strconv.Itoa(i)
		i++
		_, ok = r.Containers[containerName]
	}

	if r.Containers == nil {
		r.Containers = make(map[string]Container)
	}
	r.Containers[containerName] = c

	return containerName
}

// The purpose of Prepare() is to ensure a room is properly setup before anyone looks into it or enters it
// That way if there should be anything in the room prior, it will already be there.
// For example, mobs shouldn't ENTER the room right as the player arrives, they should already be there.
func (r *Room) Prepare(checkAdjacentRooms bool) {

	roundNow := util.GetRoundCount()

	r.Mutators.Update(roundNow)

	if len(r.Containers) > 0 {
		for k, c := range r.Containers {
			if c.DespawnRound > 0 && c.DespawnRound <= roundNow {
				r.SendText(fmt.Sprintf(`The <ansi fg="container">%s</ansi> crumbles to dust, and is gone.`, k))
				delete(r.Containers, k)
			}
		}
	}

	// First ensure any mobs that should be here are spawned
	for idx, spawnInfo := range r.SpawnInfo {

		// Make sure to clean up any instances that may be dead
		if spawnInfo.InstanceId > 0 {
			// Mob gone missing. Reset the spawn info.
			if mob := mobs.GetInstance(spawnInfo.InstanceId); mob == nil {
				spawnInfo.InstanceId = 0
				spawnInfo.DespawnedRound = roundNow
				r.SpawnInfo[idx] = spawnInfo
				continue
			}
			continue
		}

		// If a despawn was tracked, check whether the time has been reached, else skip
		if spawnInfo.DespawnedRound > 0 {

			if roundNow < gametime.GetDate(spawnInfo.DespawnedRound).AddPeriod(spawnInfo.RespawnRate) { // Not yet ready to respawn.
				continue
			}
		}

		//
		// At this point we are good to attempt respawns
		//

		// New instances needed? Spawn them
		if spawnInfo.MobId > 0 {

			forceLevel := 0

			if spawnInfo.Level > 0 {
				forceLevel = spawnInfo.Level
			} else {

				// Get the zone settings, check for scaling
				if zConfig := GetZoneConfig(r.Zone); zConfig != nil {

					if zConfig.MobAutoScale.Minimum > 0 {
						forceLevel = zConfig.GenerateRandomLevel()
					}

					if forceLevel > 0 {
						forceLevel += spawnInfo.LevelMod
						if forceLevel < 1 {
							forceLevel = 1
						}
					}

				}
			}

			if mob := mobs.NewMobById(mobs.MobId(spawnInfo.MobId), r.RoomId, forceLevel); mob != nil {

				// If a merchant, fill up stocks on first time being loaded in
				if mob.HasShop() {
					mob.Character.Shop.Restock()
				}

				if len(spawnInfo.BuffIds) > 0 {
					mob.Character.SetPermaBuffs(spawnInfo.BuffIds)
				}

				// If there are idle commands for this spawn, overwrite.
				if len(spawnInfo.IdleCommands) > 0 {
					mob.IdleCommands = append([]string{}, spawnInfo.IdleCommands...)
				}

				if len(spawnInfo.ScriptTag) > 0 {
					mob.ScriptTag = spawnInfo.ScriptTag
				}

				if len(spawnInfo.QuestFlags) > 0 {
					mob.QuestFlags = spawnInfo.QuestFlags
				}

				// Does this mob have a special name?
				if len(spawnInfo.Name) > 0 {
					mob.Character.Name = spawnInfo.Name
				}

				if spawnInfo.ForceHostile {
					mob.Hostile = true
				}

				if spawnInfo.MaxWander != 0 {
					mob.MaxWander = spawnInfo.MaxWander
				}

				mob.Character.Zone = r.Zone
				mob.Validate()

				r.mobs = append(r.mobs, mob.InstanceId)

				spawnInfo.InstanceId = mob.InstanceId
				spawnInfo.DespawnedRound = 0

				r.SpawnInfo[idx] = spawnInfo
			}

			roomManager.roomsWithMobs[r.RoomId] = len(r.mobs)

			// Since mob spanws cannot be combined with item/gold spawns, go next loop
			continue
		}

		if spawnInfo.ItemId > 0 || spawnInfo.Gold > 0 {

			// If no container specified, or the container specified exists, then spawn the item
			if spawnInfo.Container == `` {

				if _, alreadyExists := r.FindOnFloor(fmt.Sprintf(`!%d`, spawnInfo.ItemId), false); !alreadyExists {

					if item := items.New(spawnInfo.ItemId); item.ItemId != 0 {
						r.Items = append(r.Items, item) // just append to avoid a mutex double lock
					}

				}

				if r.Gold < spawnInfo.Gold {
					r.Gold = spawnInfo.Gold
				}

				spawnInfo.DespawnedRound = roundNow

				r.SpawnInfo[idx] = spawnInfo

				continue
			}

			if containerName := r.FindContainerByName(spawnInfo.Container); containerName != `` {

				container := r.Containers[containerName]

				if _, alreadyExists := container.FindItem(fmt.Sprintf(`!%d`, spawnInfo.ItemId)); !alreadyExists {
					if item := items.New(spawnInfo.ItemId); item.ItemId != 0 {
						container.AddItem(item)
					}
				}

				if container.Gold < spawnInfo.Gold {
					container.Gold = spawnInfo.Gold
				}

				r.Containers[containerName] = container

				spawnInfo.DespawnedRound = roundNow

				r.SpawnInfo[idx] = spawnInfo

			}

		}

	}

	// Reach out one more room to prepare those exit rooms
	if !checkAdjacentRooms {
		return
	}

	prepRoomIds := []int{}
	for _, exit := range r.Exits {
		if exit.RoomId == r.RoomId {
			continue
		}
		prepRoomIds = append(prepRoomIds, exit.RoomId)
	}

	for _, exitRoomId := range prepRoomIds {

		if exitRoom := LoadRoom(exitRoomId); exitRoom != nil {

			if exitRoom.PlayerCt() < 1 { // Don't prepare rooms that players are already in
				exitRoom.Prepare(false) // Don't continue checking adjacent rooms or else gets in recursion trouble
			}
		}

	}

}

func (r *Room) CleanupMobSpawns(noCooldown bool) {

	roundNow := util.GetRoundCount()
	// First ensure any mobs that should be here are spawned
	for idx, spawnInfo := range r.SpawnInfo {

		// Make sure to clean up any instances that may be dead
		if spawnInfo.InstanceId > 0 {

			if mob := mobs.GetInstance(spawnInfo.InstanceId); mob == nil {

				spawnInfo.InstanceId = 0
				if noCooldown {
					spawnInfo.DespawnedRound = 0
				} else {
					spawnInfo.DespawnedRound = roundNow
				}

			}
		}

		r.SpawnInfo[idx] = spawnInfo
	}
}

func (r *Room) AddMob(mobInstanceId int) {

	// Do before lock
	r.MarkVisited(mobInstanceId, VisitorMob)

	if mob := mobs.GetInstance(mobInstanceId); mob != nil {
		mob.Character.RoomId = r.RoomId
		mob.Character.Zone = r.Zone
	}

	r.mobs = append(r.mobs, mobInstanceId)

	roomManager.roomsWithMobs[r.RoomId] = len(r.mobs)
}

func (r *Room) RemoveMob(mobInstanceId int) {

	// Do before lock
	r.MarkVisited(mobInstanceId, VisitorMob, 1)

	mobLen := len(r.mobs)
	for i := 0; i < mobLen; i++ {
		if r.mobs[i] == mobInstanceId {
			r.mobs = append(r.mobs[:i], r.mobs[i+1:]...)
			break
		}
	}

	if len(r.mobs) < 1 {
		delete(roomManager.roomsWithMobs, r.RoomId)
	}
}

func (r *Room) AddItem(item items.Item, stash bool) {

	item.Validate()

	if stash {
		r.Stash = append(r.Stash, item)
	} else {
		r.Items = append(r.Items, item)
	}

}

func (r *Room) GetRandomExit() (exitName string, roomId int) {

	nonSecretExitCt := 0
	for _, exit := range r.Exits {
		if exit.Secret {
			continue
		}
		if exit.Lock.IsLocked() {
			continue
		}
		nonSecretExitCt++
	}

	roomSelection := util.Rand(nonSecretExitCt)

	rNow := 0
	for exitName, exit := range r.Exits {
		if exit.Secret {
			continue
		}
		if exit.Lock.IsLocked() {
			continue
		}
		if roomSelection == rNow {
			return exitName, exit.RoomId
		}
		rNow++
	}

	return ``, 0
}

func (r *Room) RemoveItem(i items.Item, stash bool) {

	if stash {
		for j := len(r.Stash) - 1; j >= 0; j-- {
			if r.Stash[j].Equals(i) {
				r.Stash = append(r.Stash[:j], r.Stash[j+1:]...)
				break
			}
		}
	} else {
		for j := len(r.Items) - 1; j >= 0; j-- {
			if r.Items[j].Equals(i) {
				r.Items = append(r.Items[:j], r.Items[j+1:]...)
				break
			}
		}
	}

}

func (r *Room) GetAllFloorItems(stash bool) []items.Item {

	found := []items.Item{}

	if stash {
		found = append(found, r.Stash...)
	}

	found = append(found, r.Items...)

	return found
}

func (r *Room) FindOnFloor(itemName string, stash bool) (items.Item, bool) {

	if stash {
		// search the stash
		closeMatchItem, matchItem := items.FindMatchIn(itemName, r.Stash...)

		if matchItem.ItemId != 0 {
			return matchItem, true
		}

		if closeMatchItem.ItemId != 0 {
			return closeMatchItem, true
		}

		return items.Item{}, false
	}

	// Search floor
	closeMatchItem, matchItem := items.FindMatchIn(itemName, r.Items...)

	if matchItem.ItemId != 0 {
		return matchItem, true
	}

	if closeMatchItem.ItemId != 0 {
		return closeMatchItem, true
	}

	return items.Item{}, false
}

func (r *Room) MarkVisited(id int, vType VisitorType, subtrackTurns ...int) {

	if r.visitors == nil {
		r.visitors = make(map[VisitorType]map[int]uint64)
	}

	if _, ok := r.visitors[vType]; !ok {
		r.visitors[vType] = make(map[int]uint64)
	}

	lastSeen := util.GetTurnCount() + uint64(visitorTrackingTimeout*configs.GetConfig().TurnsPerSecond())

	if len(subtrackTurns) > 0 {
		if uint64(subtrackTurns[0]) > lastSeen {
			lastSeen = 0
		} else {
			lastSeen -= uint64(subtrackTurns[0])
		}
	}

	r.visitors[vType][id] = lastSeen
	r.lastVisited = util.GetRoundCount()
}

func (r *Room) MobCt() int {

	return len(r.mobs)
}

func (r *Room) PlayerCt() int {

	return len(r.players)
}

func (r *Room) GetMobs(findTypes ...FindFlag) []int {

	var typeFlag FindFlag = 0
	if len(findTypes) < 1 {
		typeFlag = FindAll
	} else {
		for _, ff := range findTypes {
			typeFlag |= ff
		}
	}

	// If no filtering, just copy all mobs in the room and return it
	if typeFlag == FindAll {
		return append([]int{}, r.mobs...)
	}

	mobMatches := []int{}
	var isCharmed bool = false

	for _, mobId := range r.mobs {

		mob := mobs.GetInstance(mobId)
		if mob == nil {
			continue
		}

		if typeFlag == FindAll {
			mobMatches = append(mobMatches, mobId)
			continue
		}

		if mob.Character.Aggro != nil {
			if typeFlag&FindFightingPlayer == FindFightingPlayer && mob.Character.Aggro.UserId != 0 {
				mobMatches = append(mobMatches, mobId)
				continue
			}
			if typeFlag&FindFightingMob == FindFightingMob && mob.Character.Aggro.MobInstanceId != 0 {
				mobMatches = append(mobMatches, mobId)
				continue
			}
		}

		if typeFlag&FindNative == FindNative {
			if mob.HomeRoomId == r.RoomId {
				mobMatches = append(mobMatches, mobId)
				continue
			}
			// If not native, and that was all we were looking for, abort further tests
			if typeFlag == FindNative {
				continue
			}
		}

		if typeFlag&FindHasLight == FindHasLight && mob.Character.HasBuffFlag(buffs.EmitsLight) {
			mobMatches = append(mobMatches, mobId)
			continue
		}

		// Useful to find any mobs that will always attack players
		if mob.Hostile && typeFlag&FindHostile == FindHostile {
			mobMatches = append(mobMatches, mobId)
			continue
		}

		isCharmed = mob.Character.IsCharmed()

		if isCharmed && typeFlag&FindCharmed == FindCharmed {
			mobMatches = append(mobMatches, mobId)
			continue
		}

		// If not allied with players
		// and not current aggressive to anything
		// and won't automatically attack players
		if typeFlag&FindNeutral == FindNeutral && !isCharmed && mob.Character.Aggro == nil && !mob.Hostile {
			mobMatches = append(mobMatches, mobId)
			continue
		}

		if typeFlag&FindMerchant == FindMerchant && mob.HasShop() {
			mobMatches = append(mobMatches, mobId)
			continue
		}

		if typeFlag&FindDowned == FindDowned && mob.Character.Health < 1 {
			mobMatches = append(mobMatches, mobId)
			continue
		}

		if typeFlag&FindBuffed == FindBuffed && len(mob.Character.Buffs.List) > 0 {
			mobMatches = append(mobMatches, mobId)
			continue
		}

		if typeFlag&FindHasPet == FindHasPet && mob.Character.Pet.Exists() {
			mobMatches = append(mobMatches, mobId)
			continue
		}
	}

	return mobMatches
}

func (r *Room) GetPlayers(findTypes ...FindFlag) []int {

	var typeFlag FindFlag = 0
	if len(findTypes) < 1 {
		typeFlag = FindAll
	} else {
		for _, ff := range findTypes {
			typeFlag |= ff
		}
	}

	// If no filtering, just copy all mobs in the room and return it
	if typeFlag == FindAll {
		return append([]int{}, r.players...)
	}

	playerMatches := []int{}
	var isCharmed bool = false

	for _, userId := range r.players {

		user := users.GetByUserId(userId)
		if user == nil {
			continue
		}

		if typeFlag == FindAll {
			playerMatches = append(playerMatches, userId)
			continue
		}

		if user.Character.Aggro != nil {
			if typeFlag&FindFightingPlayer == FindFightingPlayer && user.Character.Aggro.UserId != 0 {
				playerMatches = append(playerMatches, userId)
				continue
			}
			if typeFlag&FindFightingMob == FindFightingMob && user.Character.Aggro.MobInstanceId != 0 {
				playerMatches = append(playerMatches, userId)
				continue
			}
		}

		if typeFlag&FindHasLight == FindHasLight && user.Character.HasBuffFlag(buffs.EmitsLight) {
			playerMatches = append(playerMatches, userId)
			continue
		}

		isCharmed = user.Character.IsCharmed()

		if isCharmed && typeFlag&FindCharmed == FindCharmed {
			playerMatches = append(playerMatches, userId)
			continue
		}

		// If not allied with players
		// and not current aggressive to anything
		// and won't automatically attack players
		if typeFlag&FindNeutral == FindNeutral && !isCharmed && user.Character.Aggro == nil {
			playerMatches = append(playerMatches, userId)
			continue
		}

		if typeFlag&FindMerchant == FindMerchant && user.HasShop() {
			playerMatches = append(playerMatches, userId)
			continue
		}

		if typeFlag&FindDowned == FindDowned && user.Character.Health < 1 {
			playerMatches = append(playerMatches, userId)
			continue
		}

		if typeFlag&FindBuffed == FindBuffed && len(user.Character.Buffs.List) > 0 {
			playerMatches = append(playerMatches, userId)
			continue
		}

		if typeFlag&FindHasPet == FindHasPet && user.Character.Pet.Exists() {
			playerMatches = append(playerMatches, userId)
			continue
		}
	}

	return playerMatches
}

func (r *Room) IsCalm() bool {
	return !r.ArePlayersAttacking(0) && !r.AreMobsAttacking(0)
}

func (r *Room) ArePlayersAttacking(userId int) bool {

	for _, playerId := range r.players {
		if playerId == userId {
			continue
		}
		if u := users.GetByUserId(playerId); u != nil {
			if u.Character.Aggro != nil && (userId == 0 || u.Character.Aggro.UserId == userId) {
				return true
			}
		}
	}

	return false
}

func (r *Room) AreMobsAttacking(userId int) bool {

	for _, mobId := range r.mobs {
		mob := mobs.GetInstance(mobId)
		if mob == nil {
			continue
		}
		if mob.Character.Aggro != nil && (userId == 0 || mob.Character.Aggro.UserId == userId) {
			return true
		}
	}
	return false
}

// Returns a list of recent visitors and how cold the trail is getting
func (r *Room) Visitors(vType VisitorType) map[int]float64 {

	ret := make(map[int]float64)

	if _, ok := r.visitors[vType]; ok {
		for userId, expires := range r.visitors[vType] {
			ret[userId] = float64(expires-util.GetTurnCount()) / float64(visitorTrackingTimeout*configs.GetConfig().TurnsPerSecond())
		}

	}

	return ret
}

func (r *Room) HasVisited(id int, vType VisitorType) bool {
	//	r.PruneVisitors()

	if _, ok := r.visitors[vType]; !ok {
		return false
	}

	_, ok := r.visitors[vType][id]

	return ok
}

func (r *Room) GetDescriptionFormatted(lineSplit int, highlightNouns bool) string {

	desc := util.SplitStringNL(r.GetDescription(), 80)

	if highlightNouns {
		for noun, _ := range r.Nouns {
			desc = strings.ReplaceAll(desc, noun, fmt.Sprintf("<ansi fg=\"187\">%s</ansi>", noun))
		}
	}

	return desc
}

func (r *Room) GetDescription() string {

	if !strings.HasPrefix(r.Description, `h:`) {
		return r.Description
	}
	hash := strings.TrimPrefix(r.Description, `h:`)

	return roomManager.roomDescriptionCache[hash]
}

func (r *Room) HasRecentVisitors() bool {

	return r.visitors != nil && len(r.visitors) > 0
}

func (r *Room) GetPublicSigns() []Sign {

	visibleSigns := []Sign{}
	for _, sign := range r.Signs {
		if sign.VisibleUserId == 0 {
			visibleSigns = append(visibleSigns, sign)
		}
	}

	return visibleSigns
}

func (r *Room) GetPrivateSigns() []Sign {

	privateSigns := []Sign{}
	for _, sign := range r.Signs {
		if sign.VisibleUserId != 0 {
			privateSigns = append(privateSigns, sign)
		}
	}

	return privateSigns
}

// Returns true if a sign was replaced
func (r *Room) AddSign(displayText string, visibleUserId int, daysBeforeDecay int) bool {

	s := Sign{
		VisibleUserId: visibleUserId,
		DisplayText:   displayText,
		Expires:       time.Now().Add(time.Hour * 24 * time.Duration(daysBeforeDecay)),
	}

	// If it's a public sign and one exists, replace it.
	// If it's a private rune and one exists for this player, replace it.
	for i, sign := range r.Signs {
		if sign.VisibleUserId == visibleUserId {
			r.Signs[i] = s
			return true
		}
	}

	r.Signs = append(r.Signs, s)
	return false
}

func (r *Room) FindByName(searchName string, findTypes ...FindFlag) (playerId int, mobInstanceId int) {
	if len(findTypes) < 1 {
		findTypes = []FindFlag{FindAll}
	}
	mobInstanceId, _ = r.findMobByName(searchName, findTypes...)
	playerId, _ = r.findPlayerByName(searchName, findTypes...)
	return playerId, mobInstanceId
}

func (r *Room) FindByPetName(searchName string) (playerId int) {
	// Map name to display name
	petOwners := map[string]int{}
	petNames := []string{}

	for _, uId := range r.GetPlayers(FindHasPet) {
		if u := users.GetByUserId(uId); u != nil {
			petOwners[u.Character.Pet.Name] = u.UserId
			petNames = append(petNames, u.Character.Pet.Name)
		}
	}

	match, closeMatch := util.FindMatchIn(searchName, petNames...)
	if match == `` {
		if closeMatch == `` {
			return 0
		}
		return petOwners[closeMatch]
	}

	return petOwners[match]
}

func (r *Room) findPlayerByName(searchName string, findTypes ...FindFlag) (int, error) {

	if len(searchName) > 1 {
		if searchName[0] == '#' {
			return 0, errors.New("user not found")
		}
		if searchName[0] == '@' {
			userIdMatch, _ := strconv.Atoi(searchName[1:])

			for _, uId := range r.GetPlayers(findTypes...) {

				if userIdMatch > 0 {
					if uId != userIdMatch {
						continue
					}
					return uId, nil
				}
			}
			return 0, errors.New("user not found")
		}
	}

	namesInRoom := []string{}
	// are they looking at a player?
	playerLookup := map[string]int{}
	for _, uId := range r.GetPlayers(findTypes...) {
		u := users.GetByUserId(uId)
		playerLookup[u.Character.Name] = u.UserId
		namesInRoom = append(namesInRoom, u.Character.Name)
	}

	closeMatch, fullMatch := util.FindMatchIn(searchName, namesInRoom...)

	if len(fullMatch) == 0 {
		fullMatch = closeMatch
	}

	if len(fullMatch) == 0 {
		return 0, errors.New("player not found")
	}

	return playerLookup[fullMatch], nil
}

func (r *Room) AddEffect(eType EffectType) bool {

	if r.Effects == nil {
		r.Effects = map[EffectType]AreaEffect{}
	}

	if _, ok := r.Effects[eType]; ok {
		return false
	}

	r.Effects[eType] = NewEffect(eType)
	return true
}

func (r *Room) HasEffect(eType EffectType) bool {

	if r.Effects == nil {
		return false
	}

	if effect, ok := r.Effects[eType]; ok {
		return !effect.Expired()
	}

	return false
}

func (r *Room) GetEffects() []AreaEffect {

	retFx := []AreaEffect{}

	if r.Effects == nil {
		return retFx
	}

	for _, details := range r.Effects {
		if details.Cooling() {
			continue
		}
		retFx = append(retFx, details)
	}

	return retFx
}

func (r *Room) RemoveEffect(eType EffectType) {
	delete(r.Effects, eType)
}

func (r *Room) findMobByName(searchName string, findTypes ...FindFlag) (int, error) {

	if len(searchName) > 1 {
		if searchName[0] == '@' {
			return 0, errors.New("mob not found")
		}
		if searchName[0] == '#' {
			mobIdMatch, _ := strconv.Atoi(searchName[1:])

			for _, mId := range r.GetMobs(findTypes...) {

				if mobIdMatch > 0 {
					if mId != mobIdMatch {
						continue
					}
					return mId, nil
				}
			}
			return 0, errors.New("mob not found")
		}
	}

	namesInRoom := []string{}
	friendlyMobs := map[int]*mobs.Mob{}
	mobLookup := map[string]int{}
	for _, mId := range r.GetMobs(findTypes...) {

		m := mobs.GetInstance(mId)

		if m.Character.IsCharmed() {
			friendlyMobs[mId] = m // Put friendly mobs at the end of the list.
			continue
		}

		mobName := fmt.Sprintf(`%s#%d`, m.Character.Name, len(namesInRoom)+1) // skeleton#1, skeleton#2 etc
		mobLookup[mobName] = mId
		namesInRoom = append(namesInRoom, mobName)

	}

	// Now add the friendly mobs (at the end)
	for mId, m := range friendlyMobs {
		mobName := fmt.Sprintf(`%s#%d`, m.Character.Name, len(namesInRoom)+1)
		mobLookup[mobName] = mId
		namesInRoom = append(namesInRoom, mobName)
		delete(friendlyMobs, mId)
	}

	closeMatch, fullMatch := util.FindMatchIn(searchName, namesInRoom...)

	if len(fullMatch) == 0 {
		fullMatch = closeMatch
	}

	if len(fullMatch) == 0 {
		return 0, errors.New("mob not found")
	}

	return mobLookup[fullMatch], nil

}

// Returns exitName, RoomExit
func (r *Room) FindExitTo(roomId int) string {

	for exitName, exit := range r.Exits {
		if exit.RoomId == roomId {
			return exitName
		}
	}

	for _, exit := range r.ExitsTemp {
		if exit.RoomId == roomId {
			return exit.Title
		}
	}

	return ""
}

func (r *Room) FindContainerByName(containerNameSearch string) string {

	if len(r.Containers) == 0 {
		return ``
	}

	containerNames := []string{}
	for containerName, _ := range r.Containers {
		containerNames = append(containerNames, containerName)
	}

	exactMatch, closeMatch := util.FindMatchIn(containerNameSearch, containerNames...)

	if len(exactMatch) > 0 {
		return exactMatch
	}

	return closeMatch
}

func (r *Room) FindNoun(noun string) (foundNoun string, nounDescription string) {

	for _, newNoun := range strings.Split(noun, ` `) {

		if desc, ok := r.Nouns[newNoun]; ok {
			if desc[0:1] == `:` {
				return desc[1:], r.Nouns[desc[1:]]
			}
			return noun, desc
		}

		if len(newNoun) < 2 {
			continue
		}

		// If ended in `s`, strip it and add a new word to the search list
		if noun[len(newNoun)-1:] == `s` {
			testNoun := newNoun[:len(newNoun)-1]
			if desc, ok := r.Nouns[testNoun]; ok {
				if desc[0:1] == `:` {
					return desc[1:], r.Nouns[desc[1:]]
				}
				return testNoun, desc
			}
		} else {
			testNoun := newNoun + `s`
			if desc, ok := r.Nouns[testNoun]; ok { // `s`` at end
				if desc[0:1] == `:` {
					return desc[1:], r.Nouns[desc[1:]]
				}
				return testNoun, desc
			}
		}

		// Switch ending of `y` to `ies`
		if noun[len(newNoun)-1:] == `y` {
			testNoun := newNoun[:len(newNoun)-1] + `ies`
			if desc, ok := r.Nouns[testNoun]; ok { // `ies` instead of `y` at end
				if desc[0:1] == `:` {
					return desc[1:], r.Nouns[desc[1:]]
				}
				return testNoun, desc
			}
		}

		if len(newNoun) < 3 {
			continue
		}

		// Strip 'es' such as 'torches'
		if noun[len(newNoun)-2:] == `es` {
			testNoun := newNoun[:len(newNoun)-2]
			if desc, ok := r.Nouns[testNoun]; ok {
				if desc[0:1] == `:` {
					return desc[1:], r.Nouns[desc[1:]]
				}
				return testNoun, desc
			}
		} else {
			testNoun := newNoun + `es`
			if desc, ok := r.Nouns[testNoun]; ok { // `es` at end
				if desc[0:1] == `:` {
					return desc[1:], r.Nouns[desc[1:]]
				}
				return testNoun, desc
			}
		}

		if len(newNoun) < 4 {
			continue
		}

		// Strip 'es' such as 'torches'
		if noun[len(newNoun)-3:] == `ies` {
			testNoun := newNoun[:len(newNoun)-3] + `y`
			if desc, ok := r.Nouns[testNoun]; ok { // `y` instead of `ies` at end
				if desc[0:1] == `:` {
					return desc[1:], r.Nouns[desc[1:]]
				}
				return testNoun, desc
			}
		}

	}

	return ``, ``
}

func (r *Room) FindExitByName(exitNameSearch string) (exitName string, exitRoomId int) {

	exitNames := []string{}
	for exitName, _ := range r.Exits {
		exitNames = append(exitNames, exitName)
	}

	for exitName, _ := range r.ExitsTemp {
		exitNames = append(exitNames, exitName)
	}

	exactMatch, closeMatch := util.FindMatchIn(exitNameSearch, exitNames...)

	if len(exactMatch) == 0 {

		exactMatchesRequired := []string{
			`southeast`, `southwest`,
			`northeast`, `northwest`,
		}
		// Do not allow prefix matches on southwest etc
		for _, requiredCloseMatchTerm := range exactMatchesRequired {
			if requiredCloseMatchTerm == closeMatch {
				return "", 0
			}
		}

		portalStr := `portal`
		if strings.HasPrefix(closeMatch, exitNameSearch) {
			exactMatch = closeMatch
		} else if strings.Contains(closeMatch, portalStr) { // If has portal in the word, lets consider a partial match on "portal"
			if exitNameSearch == portalStr {
				exactMatch = closeMatch
			} else { // partial starting match on "portal"?
				searchLen := len(exitNameSearch)
				if searchLen <= len(portalStr) {
					if portalStr[:searchLen] == exitNameSearch {
						exactMatch = closeMatch
					}
				}
			}
		}
	}
	if len(closeMatch) == 0 {
		return "", 0
	}

	if exitInfo, ok := r.Exits[exactMatch]; ok {
		return exactMatch, exitInfo.RoomId
	}

	if exitInfo, ok := r.ExitsTemp[exactMatch]; ok {
		return exitInfo.Title, exitInfo.RoomId
	}

	return "", 0
}

func (r *Room) PruneTemporaryExits() []TemporaryRoomExit {

	prunedExits := []TemporaryRoomExit{}

	for k, v := range r.ExitsTemp {
		if v.Expires.Before(time.Now()) {
			delete(r.ExitsTemp, k)
			prunedExits = append(prunedExits, v)
		}
	}
	return prunedExits
}

func (r *Room) PruneSigns() []Sign {

	prunedSigned := []Sign{}

	signCt := len(r.Signs)
	if signCt == 0 {
		return prunedSigned
	}

	for i := signCt - 1; i >= 0; i-- {
		s := r.Signs[i]
		if s.Expires.Before(time.Now()) {
			r.Signs = append(r.Signs[:i], r.Signs[i+1:]...)
			prunedSigned = append(prunedSigned, s)
		}
	}

	return prunedSigned
}

func (r *Room) PruneVisitors() int {

	if r.visitors == nil {
		r.visitors = make(map[VisitorType]map[int]uint64)
		return 0
	}

	// Make sure whoever is here has the freshest mark.
	for _, userId := range r.players {
		if _, ok := r.visitors[VisitorUser]; ok {
			r.visitors[VisitorUser][userId] = util.GetTurnCount() + uint64(visitorTrackingTimeout*configs.GetConfig().TurnsPerSecond())
		}
	}

	for _, mobId := range r.mobs {
		if _, ok := r.visitors[VisitorMob]; ok {
			r.visitors[VisitorMob][mobId] = util.GetTurnCount() + uint64(visitorTrackingTimeout*configs.GetConfig().TurnsPerSecond())
		}
	}

	pruneCt := 0

	for vType, _ := range r.visitors {

		for id, expires := range r.visitors[vType] {
			// Check whether expires is older than now
			if expires < util.GetTurnCount() {
				delete(r.visitors[vType], id)
				pruneCt++

				if len(r.visitors[vType]) < 1 {
					delete(r.visitors, vType)
				}
			}
		}

	}
	return pruneCt
}

func (r *Room) isInRoom(mobName string, userName string) bool {

	if mobName != `` {
		for _, mobInstId := range r.mobs {
			if mob := mobs.GetInstance(mobInstId); mob != nil {
				if strings.HasPrefix(mob.Character.Name, mobName) {
					return true
				}
			}
		}
	}

	if userName != `` {
		for _, userId := range r.players {
			if user := users.GetByUserId(userId); user != nil {
				if strings.HasPrefix(user.Character.Name, userName) {
					return true
				}
			}
		}
	}

	return false

}

func (r *Room) findMobExit(mobId int, mobName string) string {

	freshestTime := float64(0)
	freshestExitName := ``

	for exitName, exitInfo := range r.Exits {

		// Skip secret exits
		if exitInfo.Secret {
			continue
		}

		exitRoom := LoadRoom(exitInfo.RoomId)
		if exitRoom == nil {
			continue
		}

		for mId, timeLeft := range exitRoom.Visitors(VisitorMob) {

			if mobId > 0 && mobId != mId {
				continue
			}

			if visitorMob := mobs.GetInstance(mId); visitorMob != nil {

				if len(mobName) > 0 && !strings.HasPrefix(visitorMob.Character.Name, mobName) {
					continue
				}

				if timeLeft > freshestTime {
					freshestTime = timeLeft
					freshestExitName = exitName
				}

			}

		}

	}

	return freshestExitName

}

func (r *Room) findUserExit(userId int, userName string) string {

	freshestTime := float64(0)
	freshestExitName := ``

	for exitName, exitInfo := range r.Exits {

		// Skip secret exits
		if exitInfo.Secret {
			continue
		}

		exitRoom := LoadRoom(exitInfo.RoomId)
		if exitRoom == nil {
			continue
		}

		for uId, timeLeft := range exitRoom.Visitors(VisitorMob) {

			if userId > 0 && userId != uId {
				continue
			}

			if visitorUser := users.GetByUserId(uId); visitorUser != nil {

				if len(userName) > 0 && !strings.HasPrefix(visitorUser.Character.Name, userName) {
					continue
				}

				if timeLeft > freshestTime {
					freshestTime = timeLeft
					freshestExitName = exitName
				}

			}

		}

	}

	return freshestExitName

}

func (r *Room) RoundTick() {

	roundNow := util.GetRoundCount()

	//
	// Apply any mutators from the zone or room
	// This will only add mutators that the player
	// doesn't already have.
	//
	r.Mutators.Update(roundNow)

	var activeMutators mutators.MutatorList
	if zoneConfig := GetZoneConfig(r.Zone); zoneConfig != nil {
		activeMutators = append(r.Mutators.GetActive(), zoneConfig.Mutators.GetActive()...)
	}
	for _, mut := range activeMutators {
		spec := mut.GetSpec()
		r.ApplyBuffIdToPlayers(spec.PlayerBuffIds...)
		r.ApplyBuffIdToMobs(spec.MobBuffIds...)
		r.ApplyBuffIdToNativeMobs(spec.NativeBuffIds...)
	}
	//
	// Done adding mutator buffs
	//

	for idx, spawnInfo := range r.SpawnInfo {

		// Make sure to clean up any instances that may be dead
		if spawnInfo.InstanceId > 0 {
			if mob := mobs.GetInstance(spawnInfo.InstanceId); mob == nil {
				spawnInfo.InstanceId = 0
				spawnInfo.DespawnedRound = roundNow
				r.SpawnInfo[idx] = spawnInfo
			}
		}

	}

	// If any players are in the room
	// Update all mobs in the room that they've seen a player
	if len(r.players) > 0 {
		for _, mobInstanceId := range r.mobs {
			if mob := mobs.GetInstance(mobInstanceId); mob != nil {
				mob.BoredomCounter = 0
			}
		}
	}
}

func (r *Room) addPlayer(userId int) int {

	r.players = append(r.players, userId)

	return len(r.players)
}

// true if found
func (r *Room) RemovePlayer(userId int) (int, bool) {

	for i, v := range r.players {
		if v == userId {
			r.players = append(r.players[:i], r.players[i+1:]...)
			return len(r.players), true
		}
	}
	return len(r.players), false
}

// Spawns an item in the room unless:
// 1. Item is already in the room
// 2. (optional) Item is currently held by someone in the room
// 3. item repeat-spawned too recently
// If containerName is provided, ony that container name will be considered
func (r *Room) RepeatSpawnItem(itemId int, roundFrequency int, containerName ...string) bool {

	roundNum := util.GetRoundCount()
	spawnKey := strconv.Itoa(itemId)

	cName := ``
	if len(containerName) > 0 {
		cName = containerName[0]
		spawnKey = cName + `-` + spawnKey
	}

	// Are we detailing with a container?
	if cName != `` {

		c, ok := r.Containers[cName]

		// Container doesn't exist? Abort.
		if !ok {

			return false
		}

		// Item in the container? Abort.
		for _, item := range c.Items {
			if item.ItemId == itemId {

				return false
			}
		}

	}

	// Check if item is already in the room
	for _, item := range r.Items {
		if item.ItemId == itemId {

			return false
		}
	}

	// Check hidden as well
	for _, item := range r.Stash {
		if item.ItemId == itemId {

			return false
		}
	}

	// unlock for further processing that will require locks

	// Check whether enough time has passed since last spawn
	if lastSpawn := r.GetTempData(spawnKey); lastSpawn != nil {
		if lastSpawn.(uint64)+uint64(roundFrequency) > roundNum {
			return false
		}
	}

	// If someone is carrying it, abort
	for _, userId := range r.GetPlayers() {

		if user := users.GetByUserId(userId); user != nil {

			for _, item := range user.Character.GetAllBackpackItems() {
				if item.ItemId == itemId {
					return false
				}
			}

			for _, item := range user.Character.GetAllWornItems() {
				if item.ItemId == itemId {
					return false
				}
			}
		}
	}

	r.SetTempData(spawnKey, roundNum)

	// Create item
	itm := items.New(itemId)

	// Add to container?
	if cName != `` {

		c := r.Containers[cName]
		c.AddItem(itm)
		r.Containers[cName] = c

	} else { // Add to room

		r.AddItem(itm, false)

	}

	return true

}

func (r *Room) Id() int {
	return r.RoomId
}

func (r *Room) Validate() error {
	if r.Title == "" {
		return errors.New("title cannot be empty")
	}
	if r.GetDescription() == "" {
		return errors.New("description cannot be empty")
	}

	if len(r.SpawnInfo) > 0 {

		for idx, sInfo := range r.SpawnInfo {

			// Make sure that mob spawns remain separately defined from item/gold spawns.
			if sInfo.MobId > 0 {
				if sInfo.ItemId > 0 || sInfo.Gold > 0 {
					return errors.New(`a given spawn info cannot have a mobid if it has gold or an item as well. Theese must be separate spawn info entries.`)
				}
			}

			// Spawn periods if left empty default to 15 minutes
			if sInfo.RespawnRate == `` {
				sInfo.RespawnRate = `15 real minutes`
				r.SpawnInfo[idx] = sInfo
			}

		}
	}

	// Validate the biome.
	if r.Biome != `` {
		if _, found := GetBiome(r.Biome); !found {
			return fmt.Errorf("invalid biome: %s", r.Biome)
		}
	}

	// Make sure all items are validated (and have uids)
	for i := range r.Items {
		r.Items[i].Validate()
	}

	for i := range r.Stash {
		r.Stash[i].Validate()
	}

	for cName, c := range r.Containers {
		for i := range c.Items {
			c.Items[i].Validate()
		}
		r.Containers[cName] = c
	}

	if r.ZoneConfig.RoomId != r.RoomId {
		r.ZoneConfig = ZoneConfig{}
	} else {

		r.ZoneConfig.Validate()

	}

	return nil
}

func (r *Room) GetMapSymbol() string {
	if newSymbol, ok := MapSymbolOverrides[r.MapSymbol]; ok {
		return newSymbol
	}
	return r.MapSymbol
}

func (r *Room) Filename() string {
	return fmt.Sprintf("%d.yaml", r.RoomId)
}

func (r *Room) Filepath() string {
	zone := ZoneNameSanitize(r.Zone)
	return util.FilePath(zone, `/`, fmt.Sprintf("%d.yaml", r.RoomId))
}

func (r *Room) GetBiome() BiomeInfo {

	if r.Biome == `` {
		if r.Zone != `` {
			r.Biome = GetZoneBiome(r.Zone)
		}
	}

	bInfo, _ := GetBiome(r.Biome)

	return bInfo
}
