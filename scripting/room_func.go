package scripting

import (
	"fmt"
	"time"

	"github.com/dop251/goja"
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/parties"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func setRoomFunctions(vm *goja.Runtime) {
	vm.Set(`GetRoom`, GetRoom)
	vm.Set(`GetMap`, GetMap)
}

type ScriptRoom struct {
	roomId     int
	roomRecord *rooms.Room
}

func (r ScriptRoom) RoomId() int {
	return r.roomId
}

func (r ScriptRoom) SetTempData(key string, value any) {
	r.roomRecord.SetTempData(key, value)
}

func (r ScriptRoom) GetTempData(key string) any {
	return r.roomRecord.GetTempData(key)
}

func (r ScriptRoom) SetPermData(key string, value any) {
	r.roomRecord.SetLongTermData(key, value)
}

func (r ScriptRoom) GetPermData(key string) any {
	return r.roomRecord.GetLongTermData(key)
}

func (r ScriptRoom) GetItems() []ScriptItem {
	itms := make([]ScriptItem, 0, 5)
	for _, item := range r.roomRecord.GetAllFloorItems(false) {
		itms = append(itms, newScriptItem(item))
	}
	return itms
}

func (r ScriptRoom) SpawnItem(itemId int, inStash bool) {
	i := items.New(itemId)
	if i.ItemId != 0 {
		r.roomRecord.AddItem(i, inStash)
	}
}

func (r ScriptRoom) GetMobs() []int {
	return r.roomRecord.GetMobs()
}

func (r ScriptRoom) GetPlayers() []int {
	return r.roomRecord.GetPlayers()
}

func (r ScriptRoom) GetContainers() []string {
	keys := []string{}
	for key, _ := range r.roomRecord.Containers {
		keys = append(keys, key)
	}
	return keys
}

func (r ScriptRoom) GetExits() []map[string]any {

	exits := []map[string]any{}

	seed := string(configs.GetConfig().Seed)
	for exitName, exitInfo := range r.roomRecord.Exits {

		exitMap := map[string]any{
			"Name":   exitName,
			"RoomId": exitInfo.RoomId,
			"Secret": exitInfo.Secret,
			"Lock":   false,
		}

		if exitInfo.HasLock() {
			lockId := fmt.Sprintf(`%d-%s`, r.roomId, exitName)
			exitMap["Lock"] = map[string]any{
				"LockId":     lockId,
				"Difficulty": exitInfo.Lock.Difficulty,
				"Sequence":   util.GetLockSequence(lockId, int(exitInfo.Lock.Difficulty), seed),
			}
		} else {
			exitMap["Lock"] = nil
		}

		exits = append(exits, exitMap)
	}

	return exits
}

// Returns a list of userIds found to have the questId
// if userIdParty is specified, will only check users in the party of the user.
func (r ScriptRoom) HasQuest(questId string, partyUserId ...int) []int {

	hasQuestUsers := []int{}

	// Only check the user and their party?
	if len(partyUserId) > 0 && partyUserId[0] > 0 {

		if party := parties.Get(partyUserId[0]); party != nil {
			for _, userId := range party.GetMembers() {
				if user := users.GetByUserId(userId); user != nil {
					if user.Character.HasQuest(questId) {
						hasQuestUsers = append(hasQuestUsers, userId)
					}
				}
			}

			return hasQuestUsers
		}

		// No party, so just check the user
		if user := users.GetByUserId(partyUserId[0]); user != nil {
			if user.Character.HasQuest(questId) {
				hasQuestUsers = append(hasQuestUsers, user.UserId)
			}
		}

		return hasQuestUsers
	}

	// Just check all players
	for _, userId := range r.roomRecord.GetPlayers() {
		if user := users.GetByUserId(userId); user != nil {
			if user.Character.HasQuest(questId) {
				hasQuestUsers = append(hasQuestUsers, userId)
			}
		}
	}

	return hasQuestUsers
}

// Returns a list of userIds found to NOT have the questId
// if userIdParty is specified, will only check users in the party of the user.
func (r ScriptRoom) MissingQuest(questId string, partyUserId ...int) []int {

	missingQuestUsers := []int{}

	// Only check the user and their party?
	if len(partyUserId) > 0 && partyUserId[0] > 0 {

		if party := parties.Get(partyUserId[0]); party != nil {
			// Check all party members
			for _, userId := range party.GetMembers() {
				if user := users.GetByUserId(userId); user != nil {
					if !user.Character.HasQuest(questId) {
						missingQuestUsers = append(missingQuestUsers, userId)
					}
				}
			}

			return missingQuestUsers
		}

		// No party, so just check the user
		if user := users.GetByUserId(partyUserId[0]); user != nil {
			if !user.Character.HasQuest(questId) {
				missingQuestUsers = append(missingQuestUsers, user.UserId)
			}
		}
		return missingQuestUsers

	}

	// Just check all players
	for _, userId := range r.roomRecord.GetPlayers() {
		if user := users.GetByUserId(userId); user != nil {
			if !user.Character.HasQuest(questId) {
				missingQuestUsers = append(missingQuestUsers, userId)
			}
		}
	}

	return missingQuestUsers
}

func (r ScriptRoom) SpawnMob(mobId int) int {

	if mob := mobs.NewMobById(mobs.MobId(mobId), r.roomId); mob != nil {

		r.roomRecord.AddMob(mob.InstanceId)

		return mob.InstanceId
	}

	return 0

}

func (r ScriptRoom) AddTemporaryExit(exitNameSimple string, exitNameFancy string, exitRoomId int, roundTTL int) bool {
	tmpExit := rooms.TemporaryRoomExit{
		RoomId:  exitRoomId,
		Title:   exitNameFancy,
		UserId:  0,
		Expires: time.Now().Add(time.Duration(configs.GetConfig().RoundsToSeconds(roundTTL)) * time.Second),
	}

	// Spawn a portal in the room that leads to the portal location
	return r.roomRecord.AddTemporaryExit(exitNameSimple, tmpExit)
}

func (r ScriptRoom) RemoveTemporaryExit(exitNameSimple string, exitNameFancy string, exitRoomId int) bool {
	tmpExit := rooms.TemporaryRoomExit{
		RoomId: exitRoomId,
		Title:  exitNameFancy,
		UserId: 0,
	}

	// Spawn a portal in the room that leads to the portal location
	return r.roomRecord.RemoveTemporaryExit(tmpExit)
}

// ////////////////////////////////////////////////////////
//
// # These functions get exported to the scripting engine
//
// ////////////////////////////////////////////////////////
func GetRoom(roomId int) *ScriptRoom {
	if room := rooms.LoadRoom(roomId); room != nil {
		return &ScriptRoom{roomId, room}
	}
	return nil
}

// mapRoomId    - Room the map is centered on
// mapSize      - wide or normal
// mapHeight	- Height of the map
// mapWidth     - Width of the map
// mapName 		- The title of the map
// showSecrets  - Include secret exits/rooms?
// mapMarkers   - A list of strings representing custom map markers:
//
//	[roomId],[symbol],[legend text]
//	1,×,Here
func GetMap(mapRoomId int, mapSize string, mapHeight int, mapWidth int, mapName string, showSecrets bool, mapMarkers ...string) string {
	// mapRoomId    - Room the map is centered on
	// mapSize      - wide or normal
	// mapHeight	- Height of the map
	// mapWidth     - Width of the map
	// mapName 		- The title of the map
	// showSecrets  - Include secret exits/rooms?
	// mapMarkers   - A list of strings representing custom map markers:
	//                [roomId],[symbol],[legend text]
	//                1,×,Here
	return rooms.GetSpecificMap(mapRoomId, mapSize, mapHeight, mapWidth, mapName, showSecrets, mapMarkers)
}
