package scripting

import (
	"github.com/dop251/goja"
	"github.com/volte6/mud/mobs"
)

func setMobFunctions(vm *goja.Runtime) {
	vm.Set(`MobGetCharacterName`, MobGetCharacterName)
}

// ////////////////////////////////////////////////////////
//
// # These functions get exported to the scripting engine
//
// ////////////////////////////////////////////////////////
func MobGetCharacterName(mobInstanceId int) string {
	if mob := mobs.GetInstance(mobInstanceId); mob != nil {
		return mob.Character.Name
	}
	return `Unknown`
}
