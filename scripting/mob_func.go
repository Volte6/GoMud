package scripting

import (
	"github.com/dop251/goja"
	"github.com/volte6/mud/mobs"
)

func setMobFunctions(vm *goja.Runtime) {
	vm.Set(`GetMob`, GetMob)
}

type ScriptMob struct {
	mobInstanceId int
	mobRecord     *mobs.Mob
}

func (m ScriptMob) MobTypeId() int {
	return int(m.mobRecord.MobId)
}

func (m ScriptMob) InstanceId() int {
	return m.mobInstanceId
}

func (m ScriptMob) GetCharacterName() string {
	return m.mobRecord.Character.Name
}

func (m ScriptMob) Command(cmd string, waitTurns ...int) {
	if len(waitTurns) < 1 {
		waitTurns = append(waitTurns, 0)
	}
	commandQueue.QueueCommand(0, m.mobInstanceId, cmd, waitTurns[0])
}

// Returns true if a mob is charmed by/friendly to a player.
// If userId is ommitted, it will return true if the mob is charmed by any player.
func (m ScriptMob) IsCharmed(userId ...int) bool {
	if len(userId) < 1 {
		return m.mobRecord.Character.IsCharmed()
	}
	return m.mobRecord.Character.IsCharmed(userId[0])
}

func (m ScriptMob) CharmSet(userId int, charmRounds int, onRevertCommand ...string) {
	if len(onRevertCommand) < 1 {
		onRevertCommand = append(onRevertCommand, ``)
	}
	m.mobRecord.Character.Charm(userId, charmRounds, onRevertCommand[0])
}

func (m ScriptMob) CharmRemove() {
	if m.mobRecord.Character.Charmed == nil {
		return
	}
	m.mobRecord.Character.RemoveCharm()
}

func (m ScriptMob) CharmExpire() {
	m.mobRecord.Character.Charmed.Expire()
}

// ////////////////////////////////////////////////////////
//
// # These functions get exported to the scripting engine
//
// ////////////////////////////////////////////////////////
func GetMob(mobInstanceId int) *ScriptMob {

	if mob := mobs.GetInstance(mobInstanceId); mob != nil {
		return &ScriptMob{mobInstanceId, mob}
	}
	return nil
}
