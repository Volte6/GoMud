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
	vm.Set(`RoomSetTempData`, RoomSetTempData)
	vm.Set(`RoomGetTempData`, RoomGetTempData)
	vm.Set(`RoomGetItems`, RoomGetItems)
	vm.Set(`RoomGetMobs`, RoomGetMobs)
	vm.Set(`RoomGetPlayers`, RoomGetPlayers)
	vm.Set(`RoomGetContainers`, RoomGetContainers)
	vm.Set(`RoomGetExits`, RoomGetExits)
	vm.Set(`RoomGetMap`, RoomGetMap)
}

// ////////////////////////////////////////////////////////
//
// # These functions get exported to the scripting engine
//
// ////////////////////////////////////////////////////////
func RoomSetTempData(roomId int, key string, value any) {
	if room := rooms.LoadRoom(roomId); room != nil {
		room.SetTempData(key, value)
	}
}

func RoomGetTempData(roomId int, key string) any {
	if room := rooms.LoadRoom(roomId); room != nil {
		return room.GetTempData(key)
	}
	return nil
}

func RoomGetItems(roomId int) []items.Item {

	room := rooms.LoadRoom(roomId)
	if room == nil {
		return []items.Item{}
	}

	return room.GetAllFloorItems(false)
}

func RoomGetMobs(roomId int) []int {

	room := rooms.LoadRoom(roomId)
	if room == nil {
		return []int{}
	}

	return room.GetMobs()
}

func RoomGetPlayers(roomId int) []int {

	room := rooms.LoadRoom(roomId)
	if room == nil {
		return []int{}
	}

	return room.GetPlayers()
}

func RoomGetContainers(roomId int) []string {

	room := rooms.LoadRoom(roomId)
	if room == nil {
		return []string{}
	}

	keys := []string{}
	for key, _ := range room.Containers {
		keys = append(keys, key)
	}

	return keys
}

func RoomGetExits(roomId int) []map[string]any {

	exits := []map[string]any{}

	room := rooms.LoadRoom(roomId)
	if room == nil {
		return exits
	}

	seed := configs.GetConfig().Seed
	for exitName, exitInfo := range room.Exits {

		exitMap := map[string]any{
			"Name":   exitName,
			"RoomId": exitInfo.RoomId,
			"Secret": exitInfo.Secret,
			"Lock":   false,
		}

		if exitInfo.HasLock() {
			lockId := fmt.Sprintf(`%d-%s`, roomId, exitName)
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
func RoomGetMap(mapRoomId int, mapSize string, mapHeight int, mapWidth int, mapName string, showSecrets bool, mapMarkers ...string) string {
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
