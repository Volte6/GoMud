# User Functions

---

[GetUser(userId int) USEROBJ ](user_func.go) - _Retrieves a USEROBJ for a given userId._

|  Argument | Explanation |
| --- | --- |
| userId | The target user id to get. |

---

[USEROBJ.UserId() int](user_func.go) - _Returns the userId of the USEROBJ._

---

[USEROBJ.SetTempData(key string, value any)](user_func.go) - _Sets temporary data for the user (Lasts until the user logs off)._

Note: This is useful for saving/retrieving data that a user can carry along to multiple room scripts.

|  Argument | Explanation |
| --- | --- |
| key | A unique identifier for the data. |
| value | What you will be saving. |

---

[USEROBJ.GetTempData(key string) any](user_func.go) - _Gets temporary data for the user._

Note: This is useful for saving/retrieving data that a user can carry along to multiple room scripts.

|  Argument | Explanation |
| --- | --- |
| key | A unique identifier for the data. |

---

[USEROBJ.GetCharacterName() string](user_func.go) - _Retrieves the name of a user._

---

[USEROBJ.GetRoomId() int](user_func.go) - _Returns the roomId a user is in._

---

[USEROBJ.HasQuest(questId string) bool](user_func.go) - _Get whether a user has a quest/progress._

|  Argument | Explanation |
| --- | --- |
| questId | The quest identifier string to check, such as `3-start`. |

---

[USEROBJ.GiveQuest(questId string)](user_func.go) - _Grants a quest or progress on a quest to a user_

|  Argument | Explanation |
| --- | --- |
| questId | The quest identifier string to give, such as `3-start`. |

---

[USEROBJ.GiveBuff(buffId int)](user_func.go) - _Grants a user a Buff_

|  Argument | Explanation |
| --- | --- |
| buffId | The ID of the buff to give them. |

---

[USEROBJ.Command(cmd string, waitTurns ...int)](user_func.go) - _Forces a user to execute a command as if they entered it_

|  Argument | Explanation |
| --- | --- |
| cmd | The command to execute such as `look west` or `say goodbye`. |
| waitTurns (optional) | The number of turns (NOT rounds) to wait before executing the command. |

---

[USEROBJ.TrainSkill(skillName string, skillLevel int)](user_func.go) - _Sets a user skill level, if it's greater than what they already have_

|  Argument | Explanation |
| --- | --- |
| skillName | The name of the skill to train, such as `map` or `backstab`. |

---

[USEROBJ.MoveRoom(destRoomId int)](user_func.go) - _Quietly moves a user to a new room_

|  Argument | Explanation |
| --- | --- |
| destRoomId | The room id to move them to. |

---

[USEROBJ.GiveItem(itemId [int/Item])](user_func.go) - _Creates an item (if itemId) or accepts an Item object and puts it in the users backpack_

|  Argument | Explanation |
| --- | --- |
| itemId | The id or item object to give them. |

---

[USEROBJ.HasBuffFlag(buffFlag string) bool](user_func.go) - _Find out if a user has a specific buff flag_

|  Argument | Explanation |
| --- | --- |
| buffFlag | The buff flag to check [see buffspec.go](../buffs/buffspec.go). |

---

[USEROBJ.HasItemId(itemId int) bool](user_func.go) - _Check whether a user has an item id in their backpack_

|  Argument | Explanation |
| --- | --- |
| itemId | The ItemId to check for. |

---

[USEROBJ.GetBackpackItems() []Item](user_func.go) - _Get a list of Item objects in the users backpack_

Note: See [ITEMS.md](ITEMS.md) for details on Item objects.

---

[USEROBJ.GetAlignment() int](user_func.go) - _Get the numeric representation of a users alignment, from -100 to 100_

---

[USEROBJ.GetAlignmentName() string](user_func.go) - _Get the name of a users alignment, from Unholy to Holy_

---

[USEROBJ.SetAlignment(newAlignment int)](user_func.go) - _Sets user the alignment to a specific level_

|  Argument | Explanation |
| --- | --- |
| newAlignment | The new alignment, from -100 to 100 |

---
