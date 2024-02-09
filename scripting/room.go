package scripting

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/dop251/goja"
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/util"
)

var (
	roomVMCache       = make(map[int]*VMWrapper)
	scriptLoadTimeout = 1000 * time.Millisecond
	scriptRoomTimeout = 10 * time.Millisecond
)

func getRoomVM(roomId int) (*VMWrapper, error) {

	if vm, ok := roomVMCache[roomId]; ok {
		roomVMCache[roomId] = vm
		if vm == nil {
			return nil, errNoScript
		}
		return vm, nil
	}

	room := rooms.LoadRoom(roomId)
	if room == nil {
		return nil, fmt.Errorf("room not found: %d", roomId)
	}

	script := room.GetScript()
	if len(script) == 0 {
		roomVMCache[roomId] = nil
		return nil, errNoScript
	}

	vm := goja.New()
	setAllScriptingFunctions(vm)

	prg, err := goja.Compile(fmt.Sprintf(`room-%d`, roomId), script, false)
	if err != nil {
		finalErr := fmt.Errorf("Compile: %w", err)
		return nil, finalErr
	}

	//
	// Run the program
	//
	tmr := time.AfterFunc(scriptRoomTimeout, func() {
		vm.Interrupt(errTimeout)
	})
	if _, err = vm.RunProgram(prg); err != nil {

		// Wrap the error
		finalErr := fmt.Errorf("RunProgram: %w", err)

		if _, ok := finalErr.(*goja.Exception); ok {
			slog.Error("JSVM", "exception", finalErr)
			return nil, finalErr
		} else if errors.Is(finalErr, errTimeout) {
			slog.Error("JSVM", "interrupted", finalErr)
			return nil, finalErr
		}

		slog.Error("JSVM", "error", finalErr)
		return nil, finalErr
	}
	vm.ClearInterrupt()
	tmr.Stop()

	//
	// Run onLoad() function
	//
	tmr = time.AfterFunc(scriptLoadTimeout, func() {
		vm.Interrupt(errTimeout)
	})
	if fn, ok := goja.AssertFunction(vm.Get(`onLoad`)); ok {
		if _, err := fn(goja.Undefined(), vm.ToValue(roomId)); err != nil {
			// Wrap the error
			finalErr := fmt.Errorf("onLoad: %w", err)

			if _, ok := finalErr.(*goja.Exception); ok {
				slog.Error("JSVM", "exception", finalErr)
				return nil, finalErr
			} else if errors.Is(finalErr, errTimeout) {
				slog.Error("JSVM", "interrupted", finalErr)
				return nil, finalErr
			}

			slog.Error("JSVM", "error", finalErr)
			return nil, finalErr
		}
	}
	vm.ClearInterrupt()
	tmr.Stop()

	vmw := newVMWrapper(vm, 100)

	roomVMCache[roomId] = vmw

	return vmw, nil
}

func clearRoomVM(roomId int) {
	delete(roomVMCache, roomId)
}

func setRoomFunctions(vm *goja.Runtime) {
	vm.Set(`RoomGetItems`, RoomGetItems)
	vm.Set(`RoomGetMobs`, RoomGetMobs)
	vm.Set(`RoomGetPlayers`, RoomGetPlayers)
	vm.Set(`RoomGetContainers`, RoomGetContainers)
	vm.Set(`RoomGetExits`, RoomGetExits)
	vm.Set(`RoomSetTempData`, RoomSetTempData)
	vm.Set(`RoomGetTempData`, RoomGetTempData)
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

func RoomGetExits(roomId int) map[string]map[string]any {

	exits := map[string]map[string]any{}

	room := rooms.LoadRoom(roomId)
	if room == nil {
		return exits
	}

	seed := configs.GetConfig().Seed
	for exitName, exitInfo := range room.Exits {

		exitMap := map[string]any{
			"RoomId": exitInfo.RoomId,
			"Secret": exitInfo.Secret,
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

		exits[exitName] = exitMap
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
