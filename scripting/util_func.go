package scripting

import (
	"github.com/dop251/goja"
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/util"
)

func setUtilFunctions(vm *goja.Runtime) {
	vm.Set(`UtilGetRoundNumber`, UtilGetRoundNumber)
	vm.Set(`UtilFindMatchIn`, UtilFindMatchIn)
	vm.Set(`UtilGetSecondsToRounds`, UtilGetSecondsToRounds)
	vm.Set(`UtilGetMinutesToRounds`, UtilGetMinutesToRounds)
	vm.Set(`UtilGetSecondsToTurns`, UtilGetSecondsToTurns)
	vm.Set(`UtilGetMinutesToTurns`, UtilGetMinutesToTurns)
}

// ////////////////////////////////////////////////////////
//
// # These functions get exported to the scripting engine
//
// ////////////////////////////////////////////////////////
func UtilGetRoundNumber() uint64 {
	return util.GetRoundCount()
}

func UtilGetSecondsToRounds(seconds int) int {
	return configs.GetConfig().SecondsToRounds(seconds)
}

func UtilGetMinutesToRounds(minutes int) int {
	return configs.GetConfig().MinutesToRounds(minutes)
}

func UtilGetSecondsToTurns(seconds int) int {
	return configs.GetConfig().SecondsToTurns(seconds)
}

func UtilGetMinutesToTurns(minutes int) int {
	return configs.GetConfig().MinutesToTurns(minutes)
}

func UtilFindMatchIn(search string, items []string) map[string]string {

	matches := map[string]string{}

	match, closeMatch := util.FindMatchIn(search, items...)
	matches["exact"] = match
	matches["close"] = closeMatch

	return matches
}
