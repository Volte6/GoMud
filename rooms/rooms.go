package rooms

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/events"
	"github.com/volte6/mud/gametime"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
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

type SpawnInfo struct {
	MobId        int      `yaml:"mobid,omitempty"`           // Mob template Id to spawn
	InstanceId   int      `yaml:"-"`                         // Mob instance Id that was spawned (tracks whether exists currently)
	Container    string   `yaml:"container,omitempty"`       // If set, any item spawned will go into the container.
	ItemId       int      `yaml:"itemid,omitempty"`          // Item template Id to spawn on the floor
	Gold         int      `yaml:"gold,omitempty"`            // How much gold to spawn on the floor
	CooldownLeft uint16   `yaml:"-"`                         // How many rounds remain before it can spawn again. Only decrements if mob no longer exists.
	Cooldown     uint16   `yaml:"cooldown,omitempty"`        // How many rounds to wait before spawning again after it is killed.
	Message      string   `yaml:"message,omitempty"`         // (optional) message to display to the room when this creature spawns, instead of a default
	Name         string   `yaml:"name,omitempty"`            // (optional) if set, will override the mob's name
	ForceHostile bool     `yaml:"forcehostile,omitempty"`    // (optional) if true, forces the mob to be hostile.
	MaxWander    int      `yaml:"maxwander,omitempty"`       // (optional) if set, will override the mob's max wander distance
	IdleCommands []string `yaml:"idlecommands,omitempty"`    // (optional) list of commands to override the default of the mob. Useful when you need a mob to be more unique.
	ScriptTag    string   `yaml:"scripttag,omitempty"`       // (optional) if set, will override the mob's script tag
	QuestFlags   []string `yaml:"questflags,omitempty,flow"` // (optional) list of quest flags to set on the mob
	BuffIds      []int    `yaml:"buffids,omitempty,flow"`    // (optional) list of buffs the mob always has active
}

type FindFlag uint16
type VisitorType string

const (
	AffectsNone   = ""
	AffectsPlayer = "player" // Does it affect only the player who triggered it?
	AffectsRoom   = "room"   // Does it affect everyone in the room?

	// Useful for finding mobs/players
	FindCharmed        FindFlag = 0b000000001 // charmed
	FindNeutral        FindFlag = 0b000000010 // Not aggro, not charmed, not Hostile
	FindFightingPlayer FindFlag = 0b000000100 // aggro vs. a player
	FindFightingMob    FindFlag = 0b000001000 // aggro vs. a mob
	FindHostile        FindFlag = 0b000010000 // will auto-attack players
	FindMerchant       FindFlag = 0b000100000 // is a merchant
	FindDowned         FindFlag = 0b001000000 // hp < 1
	FindBuffed         FindFlag = 0b010000000 // has a buff
	FindHasLight       FindFlag = 0b100000000 // has a light source
	// Combinatorial flags
	FindFighting          = FindFightingPlayer | FindFightingMob // Currently in combat (aggro)
	FindIdle              = FindCharmed | FindNeutral            // Not aggro or hostile
	FindAll      FindFlag = 0b111111111
	// Visitor types
	VisitorUser = "user"
	VisitorMob  = "mob"
)

type Sign struct {
	VisibleUserId int       // What user can see it? If 0, then everyone can see it.
	DisplayText   string    // What text to display
	Expires       time.Time // When this sign expires.
}

type GameLock struct {
	Difficulty    uint8  `yaml:"difficulty,omitempty"` // 0 - no lock. greater than zero = difficulty to unlock.
	UnlockedUntil uint64 `yaml:"-"`                    // What round it was unlocked at, when util.GetRoundCount() > UnlockedUntil, it is relocked (set to zero).
}
type Container struct {
	Lock  GameLock     `yaml:"lock,omitempty"` // 0 - no lock. greater than zero = difficulty to unlock.
	Items []items.Item `yaml:"-"`              // Don't save contents, let room prep handle. What is found inside of the chest.
	Gold  int          `yaml:"-"`              // Don't save contents, let room prep handle. How much gold is found inside of the chest.
}

func (c Container) HasLock() bool {
	return c.Lock.Difficulty > 0
}

func (c *Container) AddItem(i items.Item) {
	c.Items = append(c.Items, i)
}

func (c *Container) RemoveItem(i items.Item) {
	for j := len(c.Items) - 1; j >= 0; j-- {
		if c.Items[j].Equals(i) {
			c.Items = append(c.Items[:j], c.Items[j+1:]...)
			break
		}
	}
}

func (c *Container) FindItem(itemName string) (items.Item, bool) {

	// Search floor
	closeMatchItem, matchItem := items.FindMatchIn(itemName, c.Items...)

	if matchItem.ItemId != 0 {
		return matchItem, true
	}

	if closeMatchItem.ItemId != 0 {
		return closeMatchItem, true
	}

	return items.Item{}, false
}

func (l GameLock) IsLocked() bool {
	return l.Difficulty > 0 && l.UnlockedUntil < util.GetRoundCount()
}

func (l *GameLock) SetUnlocked() {
	if l.Difficulty > 0 && l.UnlockedUntil < util.GetRoundCount() {
		l.UnlockedUntil = util.GetRoundCount() + uint64(configs.GetConfig().MinutesToRounds(5))
	}
}

func (l *GameLock) SetLocked() {
	l.UnlockedUntil = 0
}

type Room struct {
	mutex             sync.RWMutex
	RoomId            int    // a unique numeric index of the room. Also the filename.
	Zone              string // zone is a way to partition rooms into groups. Also into folders.
	ZoneRoot          bool   `yaml:"zoneroot,omitempty"`        // Is this the root room? If transported to a zone this is the room you end up in. Also copied for new room creation.
	IsBank            bool   `yaml:"isbank,omitempty"`          // Is this a bank room? If so, players can deposit/withdraw gold here.
	IsStorage         bool   `yaml:"isstorage,omitempty"`       // Is this a storage room? If so, players can add/remove objects here.
	IsCharacterRoom   bool   `yaml:"ischaracterroom,omitempty"` // Is this a room where characters can create new characters to swap between them?
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

// There is a magic portal of Chuckles, magic portal of Henry here!
// There is a magical hole in the east wall here!
type TemporaryRoomExit struct {
	RoomId  int       // Where does it lead to?
	Title   string    // Does this exist have a special title?
	UserId  int       // Who created it?
	Expires time.Time // When will it be auto-cleaned up?
}

type RoomExit struct {
	RoomId       int
	Secret       bool     `yaml:"secret,omitempty"`
	MapDirection string   `yaml:"mapdirection,omitempty"` // Optionaly indicate the direction of this exit for mapping purposes
	Lock         GameLock `yaml:"lock,omitempty"`         // 0 - no lock. greater than zero = difficulty to unlock.
}

func (re RoomExit) HasLock() bool {
	return re.Lock.Difficulty > 0
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

// Takes a room identifying string and breaks it apart into a room id and zone.
func ParseExit(exitStr string) (roomId int, zone string) {
	if index := strings.Index(exitStr, "/"); index != -1 {
		z, rm := strings.ToLower(exitStr[index+1:]), exitStr[0:index]
		roomId, _ = strconv.Atoi(rm)
		zone = z
	} else {
		roomId, _ = strconv.Atoi(exitStr)
	}
	return roomId, zone
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
	r.mutex.Lock()
	defer r.mutex.Unlock()

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
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if r.LongTermDataStore == nil {
		r.LongTermDataStore = make(map[string]any)
	}

	if value, ok := r.LongTermDataStore[key]; ok {
		return value
	}
	return nil
}

func (r *Room) SetTempData(key string, value any) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

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
	r.mutex.RLock()
	defer r.mutex.RUnlock()

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

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Load the script into a string
	if _, err := os.Stat(scriptPath); err == nil {
		if bytes, err := os.ReadFile(scriptPath); err == nil {
			return string(bytes)
		}
	}

	return ``
}

func (r *Room) GetScriptPath() string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Load any script for the room
	return strings.Replace(roomDataFilesPath+`/`+r.Filepath(), `.yaml`, `.js`, 1)
}

func (r *Room) FindTemporaryExitByUserId(userId int) (TemporaryRoomExit, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

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
	r.mutex.Lock()
	defer r.mutex.Unlock()

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
	r.mutex.Lock()
	defer r.mutex.Unlock()

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

// The purpose of Prepare() is to ensure a room is properly setup before anyone looks into it or enters it
// That way if there should be anything in the room prior, it will already be there.
// For example, mobs shouldn't ENTER the room right as the player arrives, they should already be there.
func (r *Room) Prepare(checkAdjacentRooms bool) {

	r.mutex.Lock()
	// First ensure any mobs that should be here are spawned
	for idx, spawnInfo := range r.SpawnInfo {

		// Make sure to clean up any instances that may be dead
		if spawnInfo.InstanceId > 0 {
			if mob := mobs.GetInstance(spawnInfo.InstanceId); mob == nil {
				spawnInfo.InstanceId = 0
			}
		}

		// New instances needed? Spawn them
		if spawnInfo.CooldownLeft < 1 {

			if spawnInfo.MobId > 0 && spawnInfo.InstanceId == 0 {
				if mob := mobs.NewMobById(mobs.MobId(spawnInfo.MobId), r.RoomId); mob != nil {

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
				}

				roomManager.roomsWithMobs[r.RoomId] = len(r.mobs)
			}

			if spawnInfo.ItemId > 0 || spawnInfo.Gold > 0 {

				// If no container specified, or the container specified exists, then spawn the item
				if spawnInfo.Container == `` {

					if item := items.New(spawnInfo.ItemId); item.ItemId != 0 {
						if _, alreadyExists := r.FindOnFloor(fmt.Sprintf(`!%d`, item.ItemId), false); !alreadyExists {
							r.Items = append(r.Items, item) // just append to avoid a mutex double lock
							spawnInfo.CooldownLeft = spawnInfo.Cooldown
						}
					}

					if spawnInfo.Gold > 0 {
						if r.Gold < spawnInfo.Gold {
							r.Gold = spawnInfo.Gold
						}
					}

				} else if containerName := r.FindContainerByName(spawnInfo.Container); containerName != `` {

					container := r.Containers[containerName]

					if item := items.New(spawnInfo.ItemId); item.ItemId != 0 {
						if _, alreadyExists := container.FindItem(fmt.Sprintf(`!%d`, item.ItemId)); !alreadyExists {
							container.AddItem(item)
						}
					}

					if spawnInfo.Gold > 0 {
						if container.Gold < spawnInfo.Gold {
							container.Gold = spawnInfo.Gold
						}
					}

					r.Containers[containerName] = container
				}

				spawnInfo.CooldownLeft = spawnInfo.Cooldown
			}

		}

		r.SpawnInfo[idx] = spawnInfo
	}

	r.mutex.Unlock() // Unlock the mutex before we possible end up doing another prepare on this room. (Some exits loop back to the same room)

	// Reach out one more room to prepare those exit rooms
	if checkAdjacentRooms {

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
}

func (r *Room) CleanupMobSpawns(noCooldown bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// First ensure any mobs that should be here are spawned
	for idx, spawnInfo := range r.SpawnInfo {

		// Make sure to clean up any instances that may be dead
		if spawnInfo.InstanceId > 0 {

			if mob := mobs.GetInstance(spawnInfo.InstanceId); mob == nil {

				spawnInfo.InstanceId = 0
				if noCooldown {
					spawnInfo.CooldownLeft = 0
				} else {
					spawnInfo.CooldownLeft = spawnInfo.Cooldown
				}

			}
		}

		r.SpawnInfo[idx] = spawnInfo
	}
}

func (r *Room) AddMob(mobInstanceId int) {

	// Do before lock
	r.MarkVisited(mobInstanceId, VisitorMob)

	r.mutex.Lock()
	defer r.mutex.Unlock()

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

	r.mutex.Lock()
	defer r.mutex.Unlock()

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
	r.mutex.Lock()
	defer r.mutex.Unlock()

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
	r.mutex.Lock()
	defer r.mutex.Unlock()

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
	r.mutex.RLock()
	defer r.mutex.RUnlock()

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
	r.mutex.Lock()
	defer r.mutex.Unlock()

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
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return len(r.mobs)
}

func (r *Room) PlayerCt() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return len(r.players)
}

func (r *Room) GetMobs(findTypes ...FindFlag) []int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

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
	}

	return mobMatches
}

func (r *Room) GetPlayers(findTypes ...FindFlag) []int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

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

	}

	return playerMatches
}

func (r *Room) IsCalm() bool {
	return !r.ArePlayersAttacking(0) && !r.AreMobsAttacking(0)
}

func (r *Room) ArePlayersAttacking(userId int) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

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
	r.mutex.RLock()
	defer r.mutex.RUnlock()

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

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if _, ok := r.visitors[vType]; ok {
		for userId, expires := range r.visitors[vType] {
			ret[userId] = float64(expires-util.GetTurnCount()) / float64(visitorTrackingTimeout*configs.GetConfig().TurnsPerSecond())
		}

	}

	return ret
}

func (r *Room) HasVisited(id int, vType VisitorType) bool {
	//	r.PruneVisitors()
	r.mutex.RLock()
	defer r.mutex.RUnlock()

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
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if !strings.HasPrefix(r.Description, `h:`) {
		return r.Description
	}
	hash := strings.TrimPrefix(r.Description, `h:`)

	return roomManager.roomDescriptionCache[hash]
}

func (r *Room) HasRecentVisitors() bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.visitors != nil && len(r.visitors) > 0
}

func (r *Room) GetPublicSigns() []Sign {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	visibleSigns := []Sign{}
	for _, sign := range r.Signs {
		if sign.VisibleUserId == 0 {
			visibleSigns = append(visibleSigns, sign)
		}
	}

	return visibleSigns
}

func (r *Room) GetPrivateSigns() []Sign {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

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
	r.mutex.Lock()
	defer r.mutex.Unlock()

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

func (r *Room) findPlayerByName(searchName string, findTypes ...FindFlag) (int, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

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
	r.mutex.RLock()
	defer r.mutex.RUnlock()

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
	r.mutex.RLock()
	defer r.mutex.RUnlock()

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
	r.mutex.Lock()
	defer r.mutex.Unlock()

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
	r.mutex.Lock()
	defer r.mutex.Unlock()

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
	r.mutex.Lock()
	defer r.mutex.Unlock()

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

func (r *Room) GetRoomDetails(user *users.UserRecord) *RoomTemplateDetails {

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var roomSymbol string = r.MapSymbol
	var roomLegend string = r.MapLegend

	b := r.GetBiome()

	if b.symbol != 0 {
		roomSymbol = string(b.symbol)
	}
	if b.name != `` {
		roomLegend = b.name
	}

	details := &RoomTemplateDetails{
		VisiblePlayers: []characters.FormattedName{},
		VisibleMobs:    []characters.FormattedName{},
		VisibleExits:   make(map[string]RoomExit),
		TemporaryExits: make(map[string]TemporaryRoomExit),
		Room:           r,               // The room being viewed
		UserId:         user.UserId,     // Who is viewing the room
		Character:      user.Character,  // The character of the user viewing the room
		Permission:     user.Permission, // The permission level of the user viewing the room
		RoomSymbol:     roomSymbol,
		RoomLegend:     roomLegend,
		IsDark:         b.IsDark(),
		IsNight:        gametime.IsNight(),
		IsBurning:      r.IsBurning(),
		TrackingString: ``,
	}

	tinymap := GetTinyMap(r.RoomId)

	events.AddToQueue(events.WebClientCommand{
		ConnectionId: user.ConnectionId(),
		Text:         "MODALADD:tinymap=" + strings.Join(tinymap, "\n"),
	})

	if tinyMapOn := user.GetConfigOption(`tinymap`); tinyMapOn != nil && tinyMapOn.(bool) {
		desclineWidth := 80 - 7 // 7 is the width of the tinymap
		padding := 1
		description := util.SplitString(r.GetDescription(), desclineWidth-padding)

		for i := 0; i < len(tinymap); i++ {
			if i > len(description)-1 {
				description = append(description, strings.Repeat(` `, desclineWidth))
			}

			description[i] += strings.Repeat(` `, desclineWidth-len(description[i])) + tinymap[i]
		}

		details.TinyMapDescription = strings.Join(description, "\n")
	}

	nameFlags := []characters.NameRenderFlag{}
	if user.Character.GetSkillLevel(skills.Peep) > 0 {
		nameFlags = append(nameFlags, characters.RenderHealth)
	}

	if useShortAdjectives := user.GetConfigOption(`shortadjectives`); useShortAdjectives != nil && useShortAdjectives.(bool) {
		nameFlags = append(nameFlags, characters.RenderShortAdjectives)
	}

	for _, playerId := range r.players {
		if playerId != user.UserId {

			renderFlags := append([]characters.NameRenderFlag{}, nameFlags...)

			player := users.GetByUserId(playerId)
			if player != nil {

				if player.Character.HasBuffFlag(buffs.Hidden) { // Don't show them if they are sneaking
					continue
				}

				pName := player.Character.GetPlayerName(user.UserId, renderFlags...)
				details.VisiblePlayers = append(details.VisiblePlayers, pName)
			}
		}
	}

	visibleFriendlyMobs := []characters.FormattedName{}

	for idx, mobInstanceId := range r.mobs {
		if mob := mobs.GetInstance(mobInstanceId); mob != nil {

			if mob.Character.HasBuffFlag(buffs.Hidden) { // Don't show them if they are sneaking
				continue
			}

			tmpNameFlags := nameFlags

			mobName := mob.Character.GetMobName(user.UserId, tmpNameFlags...)

			for _, qFlag := range mob.QuestFlags {
				if user.Character.HasQuest(qFlag) {
					mobName.QuestAlert = true
				}
			}

			if mob.Character.IsCharmed() {
				visibleFriendlyMobs = append(visibleFriendlyMobs, mobName)
			} else {
				details.VisibleMobs = append(details.VisibleMobs, mobName)
			}
		} else {
			r.mobs = append(r.mobs[:idx], r.mobs[idx+1:]...)
		}
	}

	// Add the friendly mobs to the end
	details.VisibleMobs = append(details.VisibleMobs, visibleFriendlyMobs...)

	for exitStr, exitInfo := range r.ExitsTemp {
		details.TemporaryExits[exitStr] = exitInfo
	}

	// Do this twice to ensure secrets are last

	for exitStr, exitInfo := range r.Exits {

		// If it's a secret room we need to make sure the player has recently been there before including it in the exits
		if exitInfo.Secret { //&& user.Permission != users.PermissionAdmin {
			if targetRm := LoadRoom(exitInfo.RoomId); targetRm != nil {
				if targetRm.HasVisited(user.UserId, VisitorUser) {
					details.VisibleExits[exitStr] = exitInfo
				}
			}
		} else {
			details.VisibleExits[exitStr] = exitInfo
		}
	}

	if searchMobName := user.Character.GetMiscData(`tracking-mob`); searchMobName != nil {

		if searchMobNameStr, ok := searchMobName.(string); ok {

			if r.isInRoom(searchMobNameStr, ``) {

				details.TrackingString = `Tracking <ansi fg="mobname">` + searchMobNameStr + `</ansi>... They are here!`
				user.Character.RemoveBuff(26)

			} else {

				allNames := []string{}

				for mobInstId, _ := range r.Visitors(VisitorMob) {
					if mob := mobs.GetInstance(mobInstId); mob != nil {
						allNames = append(allNames, mob.Character.Name)
					}
				}

				match, closeMatch := util.FindMatchIn(searchMobNameStr, allNames...)
				if match == `` && closeMatch == `` {

					details.TrackingString = `You lost the trail of <ansi fg="mobname">` + searchMobNameStr + `</ansi>`
					user.Character.RemoveBuff(26)

				} else {

					exitName := r.findMobExit(0, searchMobNameStr)
					if exitName == `` {

						details.TrackingString = `You lost the trail of <ansi fg="username">` + searchMobNameStr + `</ansi>`
						user.Character.RemoveBuff(26)

					} else {

						details.TrackingString = `Tracking <ansi fg="mobname">` + searchMobNameStr + `</ansi>... They went <ansi fg="exit">` + exitName + `</ansi>`
					}

				}
			}
		}

	}

	if searchUserName := user.Character.GetMiscData(`tracking-user`); searchUserName != nil {
		if searchUserNameStr, ok := searchUserName.(string); ok {

			if r.isInRoom(``, searchUserNameStr) {

				details.TrackingString = `Tracking <ansi fg="username">` + searchUserNameStr + `</ansi>... They are here!`
				user.Character.RemoveBuff(26)

			} else {

				allNames := []string{}

				for userId, _ := range r.Visitors(VisitorUser) {
					if u := users.GetByUserId(userId); u != nil {
						allNames = append(allNames, u.Character.Name)
					}
				}

				match, closeMatch := util.FindMatchIn(searchUserNameStr, allNames...)
				if match == `` && closeMatch == `` {

					details.TrackingString = `You lost the trail of <ansi fg="username">` + searchUserNameStr + `</ansi>`
					user.Character.RemoveBuff(26)

				} else {

					exitName := r.findUserExit(0, searchUserNameStr)
					if exitName == `` {

						details.TrackingString = `You lost the trail of <ansi fg="username">` + searchUserNameStr + `</ansi>`
						user.Character.RemoveBuff(26)

					} else {

						details.TrackingString = `Tracking <ansi fg="username">` + searchUserNameStr + `</ansi>... They went <ansi fg="exit">` + exitName + `</ansi>`
					}

				}
			}

		}
	}

	return details
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
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for idx, spawnInfo := range r.SpawnInfo {

		// Make sure to clean up any instances that may be dead
		if spawnInfo.InstanceId > 0 {
			if mob := mobs.GetInstance(spawnInfo.InstanceId); mob == nil {
				spawnInfo.InstanceId = 0
				r.SpawnInfo[idx] = spawnInfo
			}
		}

		if spawnInfo.CooldownLeft > 0 {
			spawnInfo.CooldownLeft--
			r.SpawnInfo[idx] = spawnInfo
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

	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.players = append(r.players, userId)

	return len(r.players)
}

// true if found
func (r *Room) RemovePlayer(userId int) (int, bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

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

	r.mutex.Lock()

	// Are we detailing with a container?
	if cName != `` {

		c, ok := r.Containers[cName]

		// Container doesn't exist? Abort.
		if !ok {
			r.mutex.Unlock()
			return false
		}

		// Item in the container? Abort.
		for _, item := range c.Items {
			if item.ItemId == itemId {
				r.mutex.Unlock()
				return false
			}
		}

	}

	// Check if item is already in the room
	for _, item := range r.Items {
		if item.ItemId == itemId {
			r.mutex.Unlock()
			return false
		}
	}

	// Check hidden as well
	for _, item := range r.Stash {
		if item.ItemId == itemId {
			r.mutex.Unlock()
			return false
		}
	}

	// unlock for further processing that will require locks
	r.mutex.Unlock()

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

		defaultCooldown := uint16(configs.GetConfig().MinutesToRounds(15))

		for idx, sInfo := range r.SpawnInfo {
			// Spawn periods if left empty default to 15 minutes
			if sInfo.Cooldown == 0 {
				if sInfo.MobId > 0 {
					sInfo.Cooldown = defaultCooldown
					r.SpawnInfo[idx] = sInfo
				}
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
