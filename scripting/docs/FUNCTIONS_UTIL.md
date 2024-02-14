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
