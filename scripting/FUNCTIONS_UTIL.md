# Utility Functions

---

[UtilGetRoundNumber() int](util_func.go) - _Gets the current Round number, which always counts up_

---

[UtilFindMatchIn(search string, items []any) object](util_func.go) - _Searches for a match in a list and returns a close and/or exact match. Close matches must be at least the first 3 letters of the subject_

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

---

[UtilGetSecondsToRounds(seconds int) int](util_func.go) - _Converts a number of seconds into a number of rounds_

|  Argument | Explanation |
| --- | --- |
| seconds | How many seconds you want converted into a round count. |

---

[UtilGetMinutesToRounds(minutes int) int](util_func.go) - _Converts a number of minutes into a number of rounds_

|  Argument | Explanation |
| --- | --- |
| minutes | How many minutes you want converted into a round count. |

---

[UtilGetSecondsToTurns(seconds int) int](util_func.go) - _Converts a number of seconds into a number of turns_

|  Argument | Explanation |
| --- | --- |
| seconds | How many seconds you want converted into a turn count. |

---

[UtilGetMinutesToTurns(minutes int) int](util_func.go) - _Converts a number of minutes into a number of turns_

|  Argument | Explanation |
| --- | --- |
| minutes | How many minutes you want converted into a turn count. |

---

[UtilStripPrepositions(input string) string](util_func.go) - _Strips out common prepositions and some other grammatical annoyances (such as into,to,from,the,my, etc.)_

|  Argument | Explanation |
| --- | --- |
| input | The string to strip and return. |

---
