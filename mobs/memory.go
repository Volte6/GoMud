package mobs

import (
	"github.com/volte6/mud/util"
)

func GetMemoryUsage() map[string]util.MemoryResult {

	mobMutex.Lock()
	defer mobMutex.Unlock()

	ret := map[string]util.MemoryResult{}

	ret["mobs"] = util.MemoryResult{util.MemoryUsage(mobs), len(mobs)}
	ret["allMobNames"] = util.MemoryResult{util.MemoryUsage(allMobNames), len(allMobNames)}
	ret["mobInstances"] = util.MemoryResult{util.MemoryUsage(mobInstances), len(mobInstances)}
	ret["mobsHatePlayers"] = util.MemoryResult{util.MemoryUsage(mobsHatePlayers), len(mobsHatePlayers)}

	return ret
}

func init() {
	util.AddMemoryReporter(`Mobs`, GetMemoryUsage)
}
