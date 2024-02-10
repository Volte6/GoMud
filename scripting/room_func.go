package scripting

import (
	"fmt"

	"github.com/dop251/goja"
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/rooms"
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

func (r ScriptRoom) etTempData(key string) any {
	return r.roomRecord.GetTempData(key)
}

func (r ScriptRoom) GetItems() []items.Item {
	return r.roomRecord.GetAllFloorItems(false)
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

	seed := configs.GetConfig().Seed
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
