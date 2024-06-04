# ActorObject

ActorObjects are the basic object that represents Users and NPCs

- [ActorObject](#actorobject)
  - [GetUser(userId int) ActorObject ](#getuseruserid-int-actorobject-)
  - [GetMob(mobInstanceId int) ActorObject ](#getmobmobinstanceid-int-actorobject-)
  - [ActorObject.UserId() int](#actorobjectuserid-int)
  - [ActorObject.InstanceId() int](#actorobjectinstanceid-int)
  - [ActorObject.MobTypeId() int](#actorobjectmobtypeid-int)
  - [ActorObject.SetTempData(key string, value any)](#actorobjectsettempdatakey-string-value-any)
  - [ActorObject.GetTempData(key string) any](#actorobjectgettempdatakey-string-any)
  - [ActorObject.GetCharacterName( \[wrapInTags bool\] ) string](#actorobjectgetcharactername-wrapintags-bool--string)
  - [ActorObject.GetRoomId() int](#actorobjectgetroomid-int)
  - [ActorObject.HasQuest(questId string) bool](#actorobjecthasquestquestid-string-bool)
  - [ActorObject.GiveQuest(questId string)](#actorobjectgivequestquestid-string)
  - [ActorObject.AddGold(amt int \[, bankAmt int\])](#actorobjectaddgoldamt-int--bankamt-int)
  - [ActorObject.AddHealth(amt int) int](#actorobjectaddhealthamt-int-int)
  - [ActorObject.Command(cmd string, waitTurns ...int)](#actorobjectcommandcmd-string-waitturns-int)
  - [ActorObject.TrainSkill(skillName string, skillLevel int)](#actorobjecttrainskillskillname-string-skilllevel-int)
  - [ActorObject.MoveRoom(destRoomId int)](#actorobjectmoveroomdestroomid-int)
  - [ActorObject.UpdateItem(itemId ItemObject)](#actorobjectupdateitemitemid-itemobject)
  - [ActorObject.GiveItem(itemId ItemObject)](#actorobjectgiveitemitemid-itemobject)
  - [ActorObject.TakeItem(itemId ItemObject)](#actorobjecttakeitemitemid-itemobject)
  - [ActorObject.HasBuff(buffId int) bool](#actorobjecthasbuffbuffid-int-bool)
  - [ActorObject.GiveBuff(buffId int)](#actorobjectgivebuffbuffid-int)
  - [ActorObject.HasBuffFlag(buffFlag string) bool](#actorobjecthasbuffflagbuffflag-string-bool)
  - [ActorObject.CancelBuffWithFlag(buffFlag string) bool](#actorobjectcancelbuffwithflagbuffflag-string-bool)
  - [ActorObject.ExpireBuff(buffId int)](#actorobjectexpirebuffbuffid-int)
  - [ActorObject.RemoveBuff(buffId int)](#actorobjectremovebuffbuffid-int)
  - [ActorObject.HasItemId(itemId int) bool](#actorobjecthasitemiditemid-int-bool)
  - [ActorObject.GetBackpackItems() \[\]ItemObject](#actorobjectgetbackpackitems-itemobject)
  - [ActorObject.GetAlignment() int](#actorobjectgetalignment-int)
  - [ActorObject.GetAlignmentName() string](#actorobjectgetalignmentname-string)
  - [ActorObject.ChangeAlignment(alignmentChange int)](#actorobjectchangealignmentalignmentchange-int)

## [GetUser(userId int) ActorObject ](/scripting/actor_func.go)
Retrieves a ActorObject for a given userId.

|  Argument | Explanation |
| --- | --- |
| userId | The target user id to get. |

## [GetMob(mobInstanceId int) ActorObject ](/scripting/actor_func.go)
Retrieves a ActorObject for a given mobInstanceId.

|  Argument | Explanation |
| --- | --- |
| mobInstanceId | The target mobInstanceId to get. |

## [ActorObject.UserId() int](/scripting/actor_func.go)
Returns the userId of the ActorObject.˚

_Note: Only useful for User ActorObjects - Returns zero otherwise._

## [ActorObject.InstanceId() int](/scripting/actor_func.go)
Returns the mobInstanceId of the ActorObject.

_Note: Only useful for Mob ActorObjects - Returns zero otherwise._

## [ActorObject.MobTypeId() int](/scripting/actor_func.go)
Returns the base mobId used to spawn new instances.

_Note: Only useful for Mob ActorObjects - Returns zero otherwise._

## [ActorObject.SetTempData(key string, value any)](/scripting/actor_func.go)
Sets temporary data for the ActorObject (Lasts until the ActorObject is removed from memory).

_Note: This is useful for saving/retrieving data that an ActorObject can carry along to multiple room scripts._

|  Argument | Explanation |
| --- | --- |
| key | A unique identifier for the data. |
| value | What you will be saving. If null, frees from memory. |

## [ActorObject.GetTempData(key string) any](/scripting/actor_func.go)
Gets temporary data for the ActorObject.

_Note: This is useful for saving/retrieving data that a ActorObject can carry along to multiple room scripts._

|  Argument | Explanation |
| --- | --- |
| key | A unique identifier for the data. |

## [ActorObject.GetCharacterName( [wrapInTags bool] ) string](/scripting/actor_func.go)
Retrieves the name of a ActorObject.

|  Argument | Explanation |
| --- | --- |
| wrapInTags | If true, will return the name wrapped in ansi tags with the fg set to `username` or `mobname`. |

## [ActorObject.GetRoomId() int](/scripting/actor_func.go)
Returns the roomId a ActorObject is in.

## [ActorObject.HasQuest(questId string) bool](/scripting/actor_func.go)
Get whether a ActorObject has a quest/progress.

|  Argument | Explanation |
| --- | --- |
| questId | The quest identifier string to check, such as `3-start`. |

## [ActorObject.GiveQuest(questId string)](/scripting/actor_func.go)
Grants a quest or progress on a quest to a ActorObject. If they are in a party, grants to the party members as well.

|  Argument | Explanation |
| --- | --- |
| questId | The quest identifier string to give, such as `3-start`. |

## [ActorObject.AddGold(amt int [, bankAmt int])](/scripting/actor_func.go)
Update how much gold an ActorObject has

|  Argument | Explanation |
| --- | --- |
| amt | A positive or negative amount of gold to alter the actors gold by. |
| bankAmt (optional) | A positive or negative amount of gold to alter the actors bank gold by. |

## [ActorObject.AddHealth(amt int) int](/scripting/actor_func.go)
Update how much health an ActorObject has, and returns the actual amount their health changed.

|  Argument | Explanation |
| --- | --- |
| amt | A positive or negative amount of health to alter the actors health by. |


## [ActorObject.Command(cmd string, waitTurns ...int)](/scripting/actor_func.go)
Forces an ActorObject to execute a command as if they entered it

_Note: Don't underestimate the power of this function! Complex and interesting behaviors or interactions can emerge from using it._

|  Argument | Explanation |
| --- | --- |
| cmd | The command to execute such as `look west` or `say goodbye`. |
| waitTurns (optional) | The number of turns (NOT rounds) to wait before executing the command. |

## [ActorObject.TrainSkill(skillName string, skillLevel int)](/scripting/actor_func.go)
Sets an ActorObject skill level, if it's greater than what they already have

|  Argument | Explanation |
| --- | --- |
| skillName | The name of the skill to train, such as `map` or `backstab`. |

## [ActorObject.MoveRoom(destRoomId int)](/scripting/actor_func.go)
Quietly moves an ActorObject to a new room

|  Argument | Explanation |
| --- | --- |
| destRoomId | The room id to move them to. |

## [ActorObject.UpdateItem(itemId ItemObject)](/scripting/actor_func.go)
Accepts an ItemObject to update in the players backpack. If the item does not already exist in the players backpack, it is ignored.

_Note: This is the only way to save changes made to an item in the players backpack._

|  Argument | Explanation |
| --- | --- |
| ItemObject | The item object to give them. |

## [ActorObject.GiveItem(itemId ItemObject)](/scripting/actor_func.go)
Accepts an ItemObject to put into the players backpack. This can be called multiple times to duplicate an item.

|  Argument | Explanation |
| --- | --- |
| ItemObject | The item object to give them. |

## [ActorObject.TakeItem(itemId ItemObject)](/scripting/actor_func.go)
Takes an object from the users backpack.

|  Argument | Explanation |
| --- | --- |
| ItemObject | The item object to take. |


## [ActorObject.HasBuff(buffId int) bool](/scripting/actor_func.go)
Returns true if the Actor has the buffId supplied

|  Argument | Explanation |
| --- | --- |
| buffId | The ID of the buff to look for. |

## [ActorObject.GiveBuff(buffId int)](/scripting/actor_func.go)
Grants an ActorObject a Buff

|  Argument | Explanation |
| --- | --- |
| buffId | The ID of the buff to give them. |

## [ActorObject.HasBuffFlag(buffFlag string) bool](/scripting/actor_func.go)
Find out if an ActorObject has a specific buff flag

|  Argument | Explanation |
| --- | --- |
| buffFlag | The buff flag to check [see buffspec.go](../buffs/buffspec.go). |

## [ActorObject.CancelBuffWithFlag(buffFlag string) bool](/scripting/actor_func.go)
Cancels any buffs that have the flag provided. Returns `true` if one or more were found.

|  Argument | Explanation |
| --- | --- |
| buffFlag | The buff flag to check [see buffspec.go](../buffs/buffspec.go). |

## [ActorObject.ExpireBuff(buffId int)](/scripting/actor_func.go)
Expire a buff immediately

|  Argument | Explanation |
| --- | --- |
| buffId | The ID of the buff to expire |

## [ActorObject.RemoveBuff(buffId int)](/scripting/actor_func.go)
Remove a buff without triggering onEnd

|  Argument | Explanation |
| --- | --- |
| buffId | The ID of the buff to remove |

## [ActorObject.HasItemId(itemId int) bool](/scripting/actor_func.go)
Check whether an ActorObject has an item id in their backpack

|  Argument | Explanation |
| --- | --- |
| itemId | The ItemId to check for. |

## [ActorObject.GetBackpackItems() []ItemObject](/scripting/actor_func.go)
Get a list of Item objects in the ActorObjects backpack

_Note: See [/scripting/docs/FUNCTIONS_ITEMS.md](/scripting/docs/FUNCTIONS_ITEMS.md) for details on ItemObject objects._

## [ActorObject.GetAlignment() int](/scripting/actor_func.go)
Get the numeric representation of a ActorObjects alignment, from -100 to 100

## [ActorObject.GetAlignmentName() string](/scripting/actor_func.go)
Get the name of an ActorObjects alignment, from Unholy to Holy

## [ActorObject.ChangeAlignment(alignmentChange int)](/scripting/actor_func.go)
Update the alignment by a relative amount. Caps result at -100 to 100

|  Argument | Explanation |
| --- | --- |
| alignmentChange | The alignment adjustment, from -200 to 200 |
