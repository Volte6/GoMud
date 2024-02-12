package scripting

import (
	"strings"

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
	vm.Set(`UtilStripPrepositions`, UtilStripPrepositions)
}

// ////////////////////////////////////////////////////////
//
// # These functions get exported to the scripting engine
//
// ////////////////////////////////////////////////////////
func UtilGetRoundNumber() uint64 {
	return util.GetRoundCount()
}

func UtilFindMatchIn(search string, items []string) map[string]any {

	matches := map[string]any{
		`found`: false,
		`exact`: ``,
		`close`: ``,
	}

	if len(search) == 0 {
		return matches
	}
	match, closeMatch := util.FindMatchIn(search, items...)

	// Only allow close matches that the search string is a prefix of
	if len(closeMatch) > 0 {
		if !strings.HasPrefix(closeMatch, search) {
			closeMatch = ``
		}

		if len(search) < len(closeMatch) && len(search) < 3 {
			closeMatch = ``
		}
	}

	matches["found"] = len(match) > 0 || len(closeMatch) > 0
	matches["exact"] = match
	matches["close"] = closeMatch

	return matches
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

func UtilStripPrepositions(input string) string {
	return util.StripPrepositions(input)
}
