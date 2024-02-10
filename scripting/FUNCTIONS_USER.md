# User Functions

---

[UserSetTempData(userId int, key string, value any)](user_func.go) - _Sets temporary data for the user (Lasts until the user logs off)._

Note: This is useful for saving/retrieving data that a user can carry along to multiple room scripts.

|  Argument | Explanation |
| --- | --- |
| userId | The target user to save data to. |
| key | A unique identifier for the data. |
| value | What you will be saving. |
---

[UserGetTempData(userId int, key string) any](user_func.go) - _Gets temporary data for the user._

Note: This is useful for saving/retrieving data that a user can carry along to multiple room scripts.

|  Argument | Explanation |
| --- | --- |
| userId | The target user to retrieve data from. |
| key | A unique identifier for the data. |

---

[UserGetCharacterName(userId int) string](user_func.go) - _Retrieves the name of a user._

|  Argument | Explanation |
| --- | --- |
| userId | The target user to get the name of. |

---

[UserGetRoomId(userId int) int](user_func.go) - _Returns the roomId a user is in._

|  Argument | Explanation |
| --- | --- |
| userId | The target user to get the roomId of. |

---

[UserHasQuest(userId int, questId string) bool](user_func.go) - _Get whether a user has a quest/progress._

|  Argument | Explanation |
| --- | --- |
| userId | The target user to get the status of. |
| questId | The quest identifier string to check, such as `3-start`. |

---

[UserGiveQuest(userId int, questId string)](user_func.go) - _Grants a quest or progress on a quest to a user_

|  Argument | Explanation |
| --- | --- |
| userId | The target user. |
| questId | The quest identifier string to give, such as `3-start`. |

---

[UserGiveBuff(userId int, buffId int)](user_func.go) - _Grants a user a Buff_

|  Argument | Explanation |
| --- | --- |
| userId | The target user. |
| buffId | The ID of the buff to give them. |

---

[UserCommand(userId int, cmd string, waitTurns ...int)](user_func.go) - _Forces a user to execute a command as if they entered it_

|  Argument | Explanation |
| --- | --- |
| userId | The target user. |
| cmd | The command to execute such as `look west` or `say goodbye`. |
| waitTurns (optional) | The number of turns (NOT rounds) to wait before executing the command. |

---

[UserTrainSkill(userId int, skillName string, skillLevel int)](user_func.go) - _Sets a user skill level, if it's greater than what they already have_

|  Argument | Explanation |
| --- | --- |
| userId | The target user. |
| skillName | The name of the skill to train, such as `map` or `backstab`. |

---

[UserMoveRoom(userId int, destRoomId int)](user_func.go) - _Quietly moves a user to a new room_

|  Argument | Explanation |
| --- | --- |
| userId | The target user. |
| destRoomId | The room id to move them to. |

---

[UserGiveItem(userId int, itemId int)](user_func.go) - _Creates an item by itemId and puts it in the users backpack_

|  Argument | Explanation |
| --- | --- |
| userId | The target user. |
| itemId | The ItemId to give them. |

---

[UserHasBuffFlag(userId int, buffFlag string) bool](user_func.go) - _Find out if a user has a specific buff flag_

|  Argument | Explanation |
| --- | --- |
| userId | The target user. |
| buffFlag | The buff flag to check [see buffspec.go](../buffs/buffspec.go). |

---

[UserHasItemId(userId int, itemId int) bool](user_func.go) - _Check whether a user has an item id in their backpack_

|  Argument | Explanation |
| --- | --- |
| userId | The target user. |
| itemId | The ItemId to check for. |

---

[UserGetBackpackItems(userId int) []items.Item](user_func.go) - _Get a list of items in the users backpack_

Note: See [ITEMS.md](ITEMS.md) for details on Item objects.

|  Argument | Explanation |
| --- | --- |
| userId | The target user. |

---
