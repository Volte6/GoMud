package scripting

import (
	"strings"

	"github.com/dop251/goja"
	"github.com/volte6/mud/colorpatterns"
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/gametime"
	"github.com/volte6/mud/users"
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
	vm.Set(`UtilDiceRoll`, UtilDiceRoll)
	vm.Set(`UtilGetTime`, UtilGetTime)
	vm.Set(`UtilGetTimeString`, UtilGetTimeString)
	vm.Set(`UtilSetTime`, UtilSetTime)
	vm.Set(`UtilSetTimeDay`, UtilSetTimeDay)
	vm.Set(`UtilSetTimeNight`, UtilSetTimeNight)
	vm.Set(`UtilIsDay`, UtilIsDay)
	vm.Set(`UtilLocateUser`, UtilLocateUser)
	vm.Set(`UtilApplyColorPattern`, UtilApplyColorPattern)
	vm.Set(`UtilGetConfig`, UtilGetConfig)
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

func UtilDiceRoll(diceQty int, diceSides int) int {
	return util.RollDice(diceQty, diceSides)
}

func UtilGetTime() gametime.GameDate {
	return gametime.GetDate()
}

func UtilGetTimeString() string {
	gd := gametime.GetDate()
	return gd.String()
}

func UtilSetTimeDay() {
	gametime.SetToDay(-1)
}

func UtilSetTimeNight() {
	gametime.SetToNight(-1)
}

func UtilSetTime(hour int, minutes int) {
	gametime.SetTime(hour, minutes)
}

func UtilIsDay() bool {
	return !gametime.IsNight()
}

func UtilLocateUser(idOrName any) int {

	// check if is string
	if userName, ok := idOrName.(string); ok { // handle string
		if locateUser := users.GetByCharacterName(userName); locateUser != nil {
			return locateUser.Character.RoomId
		}
	} else if userId, ok := idOrName.(int); ok { // handle int
		if user := users.GetByUserId(userId); user != nil {
			return user.Character.RoomId
		}
	}

	return 0
}

func UtilApplyColorPattern(input string, patternName string, wordsOnly ...bool) string {

	if len(wordsOnly) > 0 && wordsOnly[0] {
		return colorpatterns.ApplyColorPattern(input, patternName, colorpatterns.Words)
	}
	return colorpatterns.ApplyColorPattern(input, patternName)
}

func UtilGetConfig() configs.Config {
	return configs.GetConfig()
}
