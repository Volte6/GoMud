# Utility Functions

General purpose global functions

- [Utility Functions](#utility-functions)
  - [UtilGetRoundNumber() int](#utilgetroundnumber-int)
  - [UtilFindMatchIn(search string, items \[\]any) object](#utilfindmatchinsearch-string-items-any-object)
  - [UtilGetSecondsToRounds(seconds int) int](#utilgetsecondstoroundsseconds-int-int)
  - [UtilGetMinutesToRounds(minutes int) int](#utilgetminutestoroundsminutes-int-int)
  - [UtilGetSecondsToTurns(seconds int) int](#utilgetsecondstoturnsseconds-int-int)
  - [UtilGetMinutesToTurns(minutes int) int](#utilgetminutestoturnsminutes-int-int)
  - [UtilStripPrepositions(input string) string](#utilstripprepositionsinput-string-string)
  - [UtilDiceRoll(diceQty int, diceSides int) int](#utildicerolldiceqty-int-dicesides-int-int)
  - [UtilGetTime() object](#utilgettime-object)
  - [UtilSetTimeDay()](#utilsettimeday)
  - [UtilSetTime(hour int, minutes int)](#utilsettimehour-int-minutes-int)
  - [UtilIsDay() bool](#utilisday-bool)
  - [UtilLocateUser(search int|string) int](#utillocateusersearch-intstring-int)
  - [UtilApplyColorPattern(input string, patternName string \[, wordsOnly bool\]) string ](#utilapplycolorpatterninput-string-patternname-string--wordsonly-bool-string-)

## [UtilGetRoundNumber() int](/scripting/util_func.go) 
_Gets the current Round number, which always counts up_

## [UtilFindMatchIn(search string, items []any) object](/scripting/util_func.go)
Searches for a match in a list and returns a close and/or exact match. Close matches must be at least the first 3 letters of the subject

|  Argument | Explanation |
| --- | --- |
| search | The text to search for. |
| items | An array of strings to search. |

The `object` has the following properties:
|  Property | Explanation |
| --- | --- |
| object.found | `true` if either an exact or close match were found. |
| object.exact | empty string or Exact matching string. |
| object.close | empty string or Close matching string. |

## [UtilGetSecondsToRounds(seconds int) int](/scripting/util_func.go)
Converts a number of seconds into a number of rounds

|  Argument | Explanation |
| --- | --- |
| seconds | How many seconds you want converted into a round count. |

## [UtilGetMinutesToRounds(minutes int) int](/scripting/util_func.go)
Converts a number of minutes into a number of rounds

|  Argument | Explanation |
| --- | --- |
| minutes | How many minutes you want converted into a round count. |

## [UtilGetSecondsToTurns(seconds int) int](/scripting/util_func.go)
Converts a number of seconds into a number of turns

|  Argument | Explanation |
| --- | --- |
| seconds | How many seconds you want converted into a turn count. |

## [UtilGetMinutesToTurns(minutes int) int](/scripting/util_func.go)
Converts a number of minutes into a number of turns

|  Argument | Explanation |
| --- | --- |
| minutes | How many minutes you want converted into a turn count. |

## [UtilStripPrepositions(input string) string](/scripting/util_func.go)
Strips out common prepositions and some other grammatical annoyances (such as into,to,from,the,my, etc.)

|  Argument | Explanation |
| --- | --- |
| input | The string to strip and return. |

## [UtilDiceRoll(diceQty int, diceSides int) int](/scripting/util_func.go)
Simulates a dice roll and returns a result.

|  Argument | Explanation |
| --- | --- |
| diceQty | How many dice to roll. |
| diceSides | How many sides on each dice. |

## [UtilGetTime() object](/scripting/util_func.go)
Returns an object with details about the current day/time

The returned `object` has the following properties:
|  Property | Explanation |
| --- | --- |
| object.Day | `int` representing how many days have passed. |
| object.Hour | `int` current hour. |
| object.Hour24 | `int` current hour in 24 hour format. |
| object.Minute | `int` current minute. |
| object.AmPm | `AM` or `PM` |
| object.Night | `true` if is it currently nighttime. |
| object.DayStart | Hour that day starts (24 hour format). |
| object.NightStart | Hour that night starts (24 hour format). |

## [UtilSetTimeDay()](/scripting/util_func.go)
Sets the time to 1 round before day breaks.

## [UtilSetTime(hour int, minutes int)](/scripting/util_func.go)
Sets the game time to a specific `hour:minutes`, in 24 hour time.

_Example: `5:30pm` would be `UtilSetTime(17, 30)`_

|  Argument | Explanation |
| --- | --- |
| hour | The hour to set to (0-23) |
| minutes | The minutes to set to (0-59) |

## [UtilIsDay() bool](/scripting/util_func.go)
Returns true if it is currently daytime.

## [UtilLocateUser(search int|string) int](/scripting/util_func.go)
Returns the roomId of the user, or 0 (zero) if not found.

|  Argument | Explanation |
| --- | --- |
| search | username or userId to find |

## [UtilApplyColorPattern(input string, patternName string [, wordsOnly bool]) string ](/scripting/util_func.go)
Applies a color pattern to a string, and returns the colorized string

|  Argument | Explanation |
| --- | --- |
| input | plain text string you want to colorize |
| patternName | the name of the color pattern you want to apply, such as "rainbow" - [see colorpatterns/colorpatterns.go](../../colorpatterns/colorpatterns.go) |
| wordsOnly | If true, colors only change on a per-word basis. |


