package users

import "github.com/volte6/gomud/util"

func GetMemoryUsage() map[string]util.MemoryResult {

	ret := map[string]util.MemoryResult{}

	ret["Users"] = util.MemoryResult{util.MemoryUsage(userManager.Users), len(userManager.Users)}
	ret["Usernames"] = util.MemoryResult{util.MemoryUsage(userManager.Usernames), len(userManager.Usernames)}
	ret["Connections"] = util.MemoryResult{util.MemoryUsage(userManager.Connections), len(userManager.Connections)}
	ret["UserConnections"] = util.MemoryResult{util.MemoryUsage(userManager.UserConnections), len(userManager.UserConnections)}

	return ret
}

func init() {
	util.AddMemoryReporter(`Users`, GetMemoryUsage)
}
