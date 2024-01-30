package items

import "github.com/volte6/mud/util"

func GetMemoryUsage() map[string]util.MemoryResult {
	ret := map[string]util.MemoryResult{}

	ret["items"] = util.MemoryResult{util.MemoryUsage(items), len(items)}

	return ret
}

func init() {
	util.AddMemoryReporter(`Items`, GetMemoryUsage)
}
