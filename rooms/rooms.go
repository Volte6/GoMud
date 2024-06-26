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
	"github.com/volte6/mud/gametime"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

const roomDataFilesPath = "_datafiles/rooms"
const visitorTrackingTimeout = 30   // 2 minutes (in rounds)
const roomUnloadTimeoutRounds = 450 // 1800 seconds (30 minutes) / 4 seconds (1 round) = 450 rounds
const defaultMapSymbol = `•`

var (
	MapSymbolOverrides = map[string]string{
		"*": defaultMapSymbol,
		//"•": "*",
	}
)

type SpawnInfo struct {
	MobId        int      `yaml:"mobid,omitempty"`        // Mob template Id to spawn
	InstanceId   int      `yaml:"-"`                      // Mob instance Id that was spawned (tracks whether exists currently)
	Container    string   `yaml:"container,omitempty"`    // If set, any item spawned will go into the container.
	ItemId       int      `yaml:"itemid,omitempty"`       // Item template Id to spawn on the floor
	Gold         int      `yaml:"gold,omitempty"`         // How much gold to spawn on the floor
	CooldownLeft uint16   `yaml:"-"`                      // How many rounds remain before it can spawn again. Only decrements if mob no longer exists.
	Cooldown     uint16   `yaml:"cooldown,omitempty"`     // How many rounds to wait before spawning again after it is killed.
	Message      string   `yaml:"message,omitempty"`      // (optional) message to display to the room when this creature spawns, instead of a default
	Name         string   `yaml:"name,omitempty"`         // (optional) if set, will override the mob's name
	ForceHostile bool     `yaml:"forcehostile,omitempty"` // (optional) if true, forces the mob to be hostile.
	MaxWander    int      `yaml:"maxwander,omitempty"`    // (optional) if set, will override the mob's max wander distance
	IdleCommands []string `yaml:"idlecommands,omitempty"` // (optional) list of commands to override the default of the mob. Useful when you need a mob to be more unique.
	ScriptTag    string   `yaml:"scripttag,omitempty"`    // (optional) if set, will override the mob's script tag
}

type FindFlag uint16

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
)

type Sign struct {
	VisibleUserId int       // What user can see it? If 0, then everyone can see it.
	DisplayText   string    // What text to display
	Expires       time.Time // When this sign expires.
}

type PropTrigger struct {
	RoomId            int    `yaml:"roomid,omitempty"`            // What room id to move them to if any
	BuffId            int    `yaml:"buffid,omitempty"`            // What buffId to apply if any
	ItemId            int    `yaml:"itemid,omitempty"`            // What item id to give them if any
	QuestToken        string `yaml:"questtoken,omitempty"`        // What quest token to give them if any
	SkillInfo         string `yaml:"skillinfo,omitempty"`         // skill to give, format: skillId:skillLevel such as "map:1"
	Affected          string `yaml:"affected,omitempty"`          // Who is affected? If nobody, then nobody. But an item id would have to drop to the floor then.
	DescriptionPlayer string `yaml:"descriptionplayer,omitempty"` // "You press the eyes of the raven, and follow a secret entrance to the west!"
	DescriptionRoom   string `yaml:"descriptionroom,omitempty"`   // "%s presses in the eyes of the raven, and falls through to the room to the west!"
	MapInfo           string `yaml:"mapinfo,omitempty"`           // roomId[/wide] - if wide, then it will be a wide map
}

type RequirementType string

const (
	RequiresQuestToken    RequirementType = "questtoken"
	RequiresNotQuestToken RequirementType = "notquesttoken"
	RequiresItemId        RequirementType = "itemid"
	RequiresBuffId        RequirementType = "buffid"
	RequiresBuffFlag      RequirementType = "buff-flag"
	RequiresNoMobs        RequirementType = "nomobs"
)

type PropRequirement struct {
	Type             RequirementType
	IdNumber         int    `yaml:"idnumber,omitempty"`
	IdString         string `yaml:"idstring,omitempty"`
	RejectionMessage string `yaml:"rejectionmessage,omitempty"`
}

type RoomProp struct {
	Nouns         []string          `yaml:"nouns,omitempty,flow"`   // A list of nouns to match in order to interact with it. For example "raven", "eyes", "bird", "raven eyes"
	Verbs         []string          `yaml:"verbs,omitempty,flow"`   // A list of verbs to match in order to trigger it, for example "touch raven", "poke eyes", "press bird"
	Requirements  []PropRequirement `yaml:"requirements,omitempty"` // A list of requirements to interact with the prop
	Cooldown      int               `yaml:"cooldown,omitempty"`     // How many seconds before the prop can be triggered again?
	Description   string            `yaml:"description,omitempty"`  // The description of the noun when looked at (if any)
	Trigger       PropTrigger       `yaml:"trigger,omitempty"`      // Details of what triggers when a verb+noun is matched
	lastTriggered time.Time         // When was the prop last triggered?
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
	ZoneRoot          bool   `yaml:"zoneroot,omitempty"`  // Is this the root room? If transported to a zone this is the room you end up in. Also copied for new room creation.
	IsBank            bool   `yaml:"isbank,omitempty"`    // Is this a bank room? If so, players can deposit/withdraw gold here.
	IsStorage         bool   `yaml:"isstorage,omitempty"` // Is this a storage room? If so, players can add/remove objects here.
	Title             string
	Description       string
	Props             []RoomProp           `yaml:"props,omitempty"`      // A list of props in the room
	MapSymbol         string               `yaml:"mapsymbol,omitempty"`  // The symbol to use when generating a map of the zone
	MapLegend         string               `yaml:"maplegend,omitempty"`  // The text to display in the legend for this room. Should be one word.
	Biome             string               `yaml:"biome,omitempty"`      // The biome of the room. Used for weather generation.
	Containers        map[string]Container `yaml:"containers,omitempty"` // If this room has a chest, what is in it?
	Exits             map[string]RoomExit
	ExitsTemp         map[string]TemporaryRoomExit `yaml:"-"` // Temporary exits that will be removed after a certain time. Don't bother saving on sever shutting down.
	Items             []items.Item                 `yaml:"items,omitempty"`
	Gold              int                          `yaml:"gold,omitempty"`              // How much gold is on the ground?
	SpawnInfo         []SpawnInfo                  `yaml:"spawninfo,omitempty"`         // key is creature ID, value is spawn chance
	SkillTraining     map[string]TrainingRange     `yaml:"skilltraining,omitempty"`     // list of skills that can be trained in this room
	Signs             []Sign                       `yaml:"sign,omitempty"`              // list of scribbles in the room
	Stash             []items.Item                 `yaml:"stash,omitempty"`             // list of items in the room that are not visible to players
	IdleMessages      []string                     `yaml:"idlemessages,omitempty"`      // list of messages that can be displayed to players in the room
	LastIdleMessage   uint8                        `yaml:"-"`                           // index of the last idle message displayed
	LongTermDataStore map[string]any               `yaml:"longtermdatastore,omitempty"` // Long term data store for the room
	players           []int                        `yaml:"-"`                           // list of user IDs currently in the room
	mobs              []int                        `yaml:"-"`                           // list of mob instance IDs currently in the room. Does not get saved.
	visitors          map[int]uint64               `yaml:"visitors,omitempty"`          // list of user IDs that have visited this room, and the last round they did
	lastVisitor       uint64                       `yaml:"-"`                           // last round a visitor was in the room
	tempDataStore     map[string]any               `yaml:"-"`                           // Temporary data store for the room
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
		players:       []int{},
		visitors:      make(map[int]uint64),
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

func (r *Room) GetNouns() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	nouns := []string{}
	if len(r.Props) < 1 {
		return nouns
	}

	for _, prop := range r.Props {
		nouns = append(nouns, prop.Nouns...)
	}

	return nouns
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

func (r *Room) GetRandomNoun() string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if len(r.Props) == 0 {
		return ""
	}

	propNouns := []string{}
	for _, prop := range r.Props {
		propNouns = append(propNouns, prop.Nouns...)
	}

	return propNouns[util.Rand(len(propNouns))]
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
					if mob.IsMerchant {
						if len(mob.ShopStock) == 0 {

							for _, cmd := range mob.IdleCommands {

								if strings.HasPrefix(cmd, "restock ") {

									cmdParts := strings.Split(cmd, " ")
									if len(cmdParts) == 2 {

										parts := strings.Split(cmdParts[1], "/")
										itemId, _ := strconv.Atoi(parts[0])
										qty := 1
										if len(parts) == 2 {
											qty, _ = strconv.Atoi(parts[1])
										}

										mob.ShopStock[itemId] = qty

									}
								}
								if strings.HasPrefix(cmd, "restockservant ") {
									cmdParts := strings.Split(cmd, " ")
									if len(cmdParts) == 2 {

										parts := strings.Split(cmdParts[1], "/")
										mobId, _ := strconv.Atoi(parts[0])

										qty := 1
										if len(parts) > 2 {
											qty, _ = strconv.Atoi(parts[1])
										}

										price := 999999
										if len(parts) > 1 {
											price, _ = strconv.Atoi(parts[2])
										}

										found := false
										for idx, servantInfo := range mob.ShopServants {
											if servantInfo.MobId == mobs.MobId(mobId) {
												mob.ShopServants[idx].Quantity = qty
												mob.ShopServants[idx].Price = price
												found = true
											}
										}

										if !found {
											mob.ShopServants = append(mob.ShopServants, mobs.MobForHire{
												MobId:    mobs.MobId(mobId),
												Quantity: qty,
												Price:    price,
											})
										}

									}
								}
							}
						}
					}

					// If there are idle commands for this spawn, overwrite.
					if len(spawnInfo.IdleCommands) > 0 {
						mob.IdleCommands = append([]string{}, spawnInfo.IdleCommands...)
					}

					if len(spawnInfo.ScriptTag) > 0 {
						mob.ScriptTag = spawnInfo.ScriptTag
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

func (r *Room) TryLookProp(cmd string, restOfInput string) (description string, ok bool) {
	// Check for a matching noun and if found, return the description
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if cmd == "look" {

		restOfInput = strings.ToLower(restOfInput)

		for _, prop := range r.Props {
			for _, noun := range prop.Nouns {
				if strings.Contains(restOfInput, noun) {
					return prop.Description, true
				}
			}
		}

	}

	return "", false
}

// Tries a trigger, and if it finds a match, returns the trigger and whether it's on cooldown.
func (r *Room) TryTrigger(cmd string, restOfInput string, userId int) (matchingTrigger *PropTrigger, onCooldown bool, rejectionMessage string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if len(r.Props) == 0 {
		return matchingTrigger, onCooldown, rejectionMessage
	}

	user := users.GetByUserId(userId)
	if user == nil {
		return matchingTrigger, onCooldown, rejectionMessage
	}

	restOfInput = strings.ToLower(restOfInput)

	onCooldown = false
	for p, prop := range r.Props {

		verbMatch := false
		if len(prop.Verbs) > 0 {
			for _, verb := range prop.Verbs {
				if cmd != verb {
					continue
				}
				verbMatch = true
				break
			}
		}

		nounMatch := true
		if len(prop.Nouns) > 0 {
			nounMatch = false
			for _, noun := range prop.Nouns {
				// Changed from contains
				if noun != `` { // empty nouns always match.
					if !strings.HasPrefix(noun, restOfInput) || restOfInput == `` { // Can't match nouns if there is no noun provided
						continue
					}
				}

				nounMatch = true
				break
			}
		}

		if verbMatch && nounMatch {

			rejected := false

			for _, req := range prop.Requirements {

				// Quest token (have or not have)
				if req.Type == RequiresQuestToken {

					rejected = !user.Character.HasQuest(req.IdString)
					if rejected {
						rejectionMessage = req.RejectionMessage
						break
					}

					continue
				}

				if req.Type == RequiresNotQuestToken {
					rejected = user.Character.HasQuest(req.IdString)
					if rejected {
						rejectionMessage = req.RejectionMessage
						break
					}

					continue
				}

				// ItemId (have or not have)
				if req.Type == RequiresItemId {

					if req.IdNumber > 0 {
						_, found := user.Character.FindInBackpack(fmt.Sprintf(`!%d`, req.IdNumber))
						if !found {
							_, found = user.Character.FindOnBody(fmt.Sprintf(`!%d`, req.IdNumber))
						}
						rejected = !found
					} else if req.IdNumber < 0 {
						_, found := user.Character.FindInBackpack(fmt.Sprintf(`!%d`, req.IdNumber*-1))
						if !found {
							_, found = user.Character.FindOnBody(fmt.Sprintf(`!%d`, req.IdNumber*-1))
						}
						rejected = found
					}

					if rejected {
						rejectionMessage = req.RejectionMessage
						break
					}

					continue
				}

				// BuffId (have or not have)
				if req.Type == RequiresBuffId {

					if req.IdNumber > 0 {
						rejected = !user.Character.HasBuff(req.IdNumber) // positive buffId means they must have this buff
					} else if req.IdNumber < 0 { // negative buffId means they must not have this buff
						rejected = user.Character.HasBuff(req.IdNumber * -1)
					}

					if rejected {
						rejectionMessage = req.RejectionMessage
						break
					}

					continue
				}

				if req.Type == RequiresBuffFlag {

					rejected = !user.Character.HasBuffFlag(buffs.Flag(req.IdString)) // positive buffId means they must have this buff

					if rejected {
						rejectionMessage = req.RejectionMessage
						break
					}

					continue
				}

				// No mobs in the room
				if req.Type == RequiresNoMobs {
					if len(r.mobs) > 0 {
						rejected = true
						rejectionMessage = req.RejectionMessage
					}

					continue
				}
			}

			if rejected {
				if len(rejectionMessage) == 0 {
					rejectionMessage = "Something prevents it..."
				}
				return matchingTrigger, onCooldown, rejectionMessage
			}

			// We have a match!
			matchingTrigger = &r.Props[p].Trigger

			// Check if the prop is on cooldown
			if prop.Cooldown > 0 {
				if time.Now().After(prop.lastTriggered.Add(time.Second * time.Duration(prop.Cooldown))) {
					r.Props[p].lastTriggered = time.Now()
				} else {
					onCooldown = true
				}
			}

			// return the match data
			return matchingTrigger, onCooldown, rejectionMessage

		}
	}

	return matchingTrigger, onCooldown, rejectionMessage
}

func (r *Room) MarkVisited(userId int) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.visitors == nil {
		r.visitors = map[int]uint64{}
	}
	r.visitors[userId] = util.GetRoundCount() + visitorTrackingTimeout
	r.lastVisitor = util.GetRoundCount()
}

func (r *Room) PlayerCt() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return len(r.players)
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

func (r *Room) MobCt() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return len(r.mobs)
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

		if typeFlag&FindMerchant == FindMerchant && mob.IsMerchant {
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
func (r *Room) Visitors() map[int]float64 {

	ret := make(map[int]float64)

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for userId, expires := range r.visitors {
		ret[userId] = float64(expires-util.GetRoundCount()) / visitorTrackingTimeout
	}

	return ret
}

func (r *Room) HasVisited(userId int) bool {
	//	r.PruneVisitors()
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	_, ok := r.visitors[userId]
	return ok
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
		return 0
	}

	// Make sure whoever is here has the freshest mark.
	for _, userId := range r.players {
		if r.visitors == nil {
			r.visitors = make(map[int]uint64)
		}
		r.visitors[userId] = util.GetRoundCount() + visitorTrackingTimeout
	}

	pruneCt := 0
	for userId, expires := range r.visitors {
		// Check whether expires is older than now
		if expires < util.GetRoundCount() {
			delete(r.visitors, userId)
			pruneCt++
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
		Nouns:          r.GetNouns(),
		IsDark:         b.IsDark(),
		IsNight:        gametime.IsNight(),
	}

	if tinyMapOn := user.GetConfigOption(`tinymap`); tinyMapOn != nil && tinyMapOn.(bool) {
		tinymap := GetTinyMap(r.RoomId)
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
				if targetRm.HasVisited(user.UserId) {
					details.VisibleExits[exitStr] = exitInfo
				}
			}
		} else {
			details.VisibleExits[exitStr] = exitInfo
		}
	}

	return details
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

	if len(r.Props) > 0 {
		for p, prop := range r.Props {
			// make all verbs and all nouns lowercase.
			for i, verb := range prop.Verbs {
				r.Props[p].Verbs[i] = strings.ToLower(verb)
			}
			for i, noun := range prop.Nouns {
				r.Props[p].Nouns[i] = strings.ToLower(noun)
			}
		}
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
