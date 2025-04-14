package items

import "github.com/GoMudEngine/GoMud/internal/util"

func GetMemoryUsage() map[string]util.MemoryResult {
	ret := map[string]util.MemoryResult{}

	ret["items"] = util.MemoryResult{util.MemoryUsage(items), len(items)}

	return ret
}

func init() {
	util.AddMemoryReporter(`Items`, GetMemoryUsage)
}
