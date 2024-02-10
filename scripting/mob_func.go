package scripting

import (
	"github.com/dop251/goja"
	"github.com/volte6/mud/mobs"
)

func setMobFunctions(vm *goja.Runtime) {
	vm.Set(`MobGetCharacterName`, MobGetCharacterName)
	vm.Set(`MobCommand`, MobCommand)
	vm.Set(`MobCharmed`, MobCharmed)
	vm.Set(`MobCharmSet`, MobCharmSet)
	vm.Set(`MobCharmRemove`, MobCharmRemove)
	vm.Set(`MobCharmExpire`, MobCharmExpire)

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

func MobCommand(mobInstanceId int, cmd string, waitTurns ...int) {
	if len(waitTurns) < 1 {
		waitTurns = append(waitTurns, 0)
	}
	commandQueue.QueueCommand(0, mobInstanceId, cmd, waitTurns[0])
}

// Returns true if a mob is charmed by/friendly to a player.
// If userId is ommitted, it will return true if the mob is charmed by any player.
func MobCharmed(mobInstanceId int, userId ...int) bool {

	if mob := mobs.GetInstance(mobInstanceId); mob != nil {
		if len(userId) < 1 {
			return mob.Character.IsCharmed()
		}
		return mob.Character.IsCharmed(userId[0])
	}
	return false
}

func MobCharmSet(mobInstanceId int, userId int, charmRounds int, onRevertCommand ...string) {
	if len(onRevertCommand) < 1 {
		onRevertCommand = append(onRevertCommand, ``)
	}
	if mob := mobs.GetInstance(mobInstanceId); mob != nil {
		mob.Character.Charm(userId, charmRounds, onRevertCommand[0])
	}
}

func MobCharmRemove(mobInstanceId int) {
	if mob := mobs.GetInstance(mobInstanceId); mob != nil {
		if mob.Character.Charmed == nil {
			return
		}
		mob.Character.RemoveCharm()
	}
}

func MobCharmExpire(mobInstanceId int) {

	if mob := mobs.GetInstance(mobInstanceId); mob != nil {
		if mob.Character.Charmed == nil {
			return
		}
		mob.Character.Charmed.Expire()
	}
}
