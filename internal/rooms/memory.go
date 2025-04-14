package rooms

import (
	"github.com/GoMudEngine/GoMud/internal/util"
)

func GetMemoryUsage() map[string]util.MemoryResult {

	ret := map[string]util.MemoryResult{}

	ret["rooms"] = util.MemoryResult{util.MemoryUsage(roomManager.rooms), len(roomManager.rooms)}
	ret["zones"] = util.MemoryResult{util.MemoryUsage(roomManager.zones), len(roomManager.zones)}
	ret["roomsWithUsers"] = util.MemoryResult{util.MemoryUsage(roomManager.roomsWithUsers), len(roomManager.roomsWithUsers)}
	ret["roomIdToFileCache"] = util.MemoryResult{util.MemoryUsage(roomManager.roomIdToFileCache), len(roomManager.roomIdToFileCache)}

	return ret
}

func init() {
	util.AddMemoryReporter(`Rooms`, GetMemoryUsage)
}
