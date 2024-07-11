# ActorObject

ActorObjects are the basic object that represents Users and NPCs

- [ActorObject](#actorobject)
  - [ActorNames(actors \[\]ActorObject) string ](#actornamesactors-actorobject-string-)
  - [GetUser(userId int) ActorObject ](#getuseruserid-int-actorobject-)
  - [GetMob(mobInstanceId int) ActorObject ](#getmobmobinstanceid-int-actorobject-)
  - [ActorObject.UserId() int](#actorobjectuserid-int)
  - [ActorObject.InstanceId() int](#actorobjectinstanceid-int)
  - [ActorObject.MobTypeId() int](#actorobjectmobtypeid-int)
  - [ActorObject.GetRace() string](#actorobjectgetrace-string)
  - [ActorObject.GetSize() string](#actorobjectgetsize-string)
  - [ActorObject.GetLevel() int](#actorobjectgetlevel-int)
  - [ActorObject.GetStat(statName string) int](#actorobjectgetstatstatname-string-int)
  - [ActorObject.SetTempData(key string, value any)](#actorobjectsettempdatakey-string-value-any)
  - [ActorObject.GetTempData(key string) any](#actorobjectgettempdatakey-string-any)
  - [ActorObject.SetMiscCharacterData(key string, value any)](#actorobjectsetmisccharacterdatakey-string-value-any)
  - [ActorObject.GetMiscCharacterData(key string) any](#actorobjectgetmisccharacterdatakey-string-any)
  - [ActorObject.GetMiscCharacterDataKeys(\[ prefix1, prefix2 \]) \[\]string](#actorobjectgetmisccharacterdatakeys-prefix1-prefix2--string)
  - [ActorObject.GetCharacterName( wrapInTags bool ) string](#actorobjectgetcharactername-wrapintags-bool--string)
  - [ActorObject.GetRoomId() int](#actorobjectgetroomid-int)
  - [ActorObject.HasQuest(questId string) bool](#actorobjecthasquestquestid-string-bool)
  - [ActorObject.GiveQuest(questId string)](#actorobjectgivequestquestid-string)
  - [ActorObject.AddGold(amt int \[, bankAmt int\])](#actorobjectaddgoldamt-int--bankamt-int)
  - [ActorObject.AddHealth(amt int) int](#actorobjectaddhealthamt-int-int)
  - [ActorObject.Command(cmd string, waitTurns ...int)](#actorobjectcommandcmd-string-waitturns-int)
  - [ActorObject.IsTameable() bool](#actorobjectistameable-bool)
  - [ActorObject.TrainSkill(skillName string, skillLevel int)](#actorobjecttrainskillskillname-string-skilllevel-int)
  - [ActorObject.GetSkillLevel(skillName string)](#actorobjectgetskilllevelskillname-string)
  - [ActorObject.MoveRoom(destRoomId int \[, leaveCharmedMobsBehind bool\] )](#actorobjectmoveroomdestroomid-int--leavecharmedmobsbehind-bool-)
  - [ActorObject.UpdateItem(itemId ItemObject)](#actorobjectupdateitemitemid-itemobject)
  - [ActorObject.GiveItem(itemId ItemObject)](#actorobjectgiveitemitemid-itemobject)
  - [ActorObject.TakeItem(itemId ItemObject)](#actorobjecttakeitemitemid-itemobject)
  - [ActorObject.HasBuff(buffId int) bool](#actorobjecthasbuffbuffid-int-bool)
  - [ActorObject.GiveBuff(buffId int)](#actorobjectgivebuffbuffid-int)
  - [ActorObject.HasBuffFlag(buffFlag string) bool](#actorobjecthasbuffflagbuffflag-string-bool)
  - [ActorObject.CancelBuffWithFlag(buffFlag string) bool](#actorobjectcancelbuffwithflagbuffflag-string-bool)
  - [ActorObject.RemoveBuff(buffId int)](#actorobjectremovebuffbuffid-int)
  - [ActorObject.HasItemId(itemId int, \[excludeWorn bool\]) bool](#actorobjecthasitemiditemid-int-excludeworn-bool-bool)
  - [ActorObject.GetBackpackItems() \[\]ItemObject](#actorobjectgetbackpackitems-itemobject)
  - [ActorObject.GetAlignment() int](#actorobjectgetalignment-int)
  - [ActorObject.GetAlignmentName() string](#actorobjectgetalignmentname-string)
  - [ActorObject.ChangeAlignment(alignmentChange int)](#actorobjectchangealignmentalignmentchange-int)
  - [ActorObject.HasSpell(spellId string)](#actorobjecthasspellspellid-string)
  - [ActorObject.LearnSpell(spellId string) bool](#actorobjectlearnspellspellid-string-bool)
  - [ActorObject.IsAggro(targetActor ActorObject)](#actorobjectisaggrotargetactor-actorobject)
  - [ActorObject.GetMobKills(mobId int) int](#actorobjectgetmobkillsmobid-int-int)
  - [ActorObject.GetRaceKills(raceName string) int](#actorobjectgetracekillsracename-string-int)
  - [ActorObject.GetHealth() int](#actorobjectgethealth-int)
  - [ActorObject.GetHealthMax() int](#actorobjectgethealthmax-int)
  - [ActorObject.GetHealthPct() float](#actorobjectgethealthpct-float)
  - [ActorObject.GetMana() int](#actorobjectgetmana-int)
  - [ActorObject.GetManaMax() int](#actorobjectgetmanamax-int)
  - [ActorObject.GetManaPct() float](#actorobjectgetmanapct-float)
  - [ActorObject.SetAdjective(adj string, addIt bool)](#actorobjectsetadjectiveadj-string-addit-bool)
  - [ActorObject.IsCharmed( \[userId1, userId2, etc \] ) bool ](#actorobjectischarmed-userid1-userid2-etc---bool-)
  - [ActorObject.CharmSet(userId int, charmRounds int, \[ onRevertCommand1, onRevertCommand2, etc \])](#actorobjectcharmsetuserid-int-charmrounds-int--onrevertcommand1-onrevertcommand2-etc-)
  - [ActorObject.CharmRemove()](#actorobjectcharmremove)
  - [ActorObject.CharmExpire()](#actorobjectcharmexpire)
  - [ActorObject.GetCharmCount() int](#actorobjectgetcharmcount-int)




## [ActorNames(actors []ActorObject) string ](/scripting/actor_func.go)
Returns a formatted list of actor names, separated by commas, then "and".

_Example: "Tim, Jim and Henry"_

|  Argument | Explanation |
| --- | --- |
| actors | An array of ActorObjects. |

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

## [ActorObject.GetRace() string](/scripting/actor_func.go)
Gets the race name of the actor, such as Human, Elf, Rodent, etc.

## [ActorObject.GetSize() string](/scripting/actor_func.go)
Returns `small`, `medium`, or `large`

## [ActorObject.GetLevel() int](/scripting/actor_func.go)
Returns the level of the actor

## [ActorObject.GetStat(statName string) int](/scripting/actor_func.go)
Returns the named stat value.

|  Argument | Explanation |
| --- | --- |
| statName | A stat name such as `strength`, `smarts`, `perception`, etc. |

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

## [ActorObject.SetMiscCharacterData(key string, value any)](/scripting/actor_func.go)
Sets permanent data for the ActorObject. 

_Note: This miscellaneous data is attached to the character data, not the user data. If the user changes characters, it will not follow._

|  Argument | Explanation |
| --- | --- |
| key | A unique identifier for the data. |
| value | What you will be saving. If null, frees from memory. |

## [ActorObject.GetMiscCharacterData(key string) any](/scripting/actor_func.go)
Gets permanent data for the ActorObject.

_Note: This miscellaneous data is attached to the character data, not the user data. If the user changes characters, it will not follow._

|  Argument | Explanation |
| --- | --- |
| key | A unique identifier for the data. |

## [ActorObject.GetMiscCharacterDataKeys([ prefix1, prefix2 ]) []string](/scripting/actor_func.go)
Gets a list of misc data keys for the ActorObject.

|  Argument | Explanation |
| --- | --- |
| prefix1, prefix2, etc | Optional strings of prefixes to return matching keys. |

## [ActorObject.GetCharacterName( wrapInTags bool ) string](/scripting/actor_func.go)
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

## [ActorObject.IsTameable() bool](/scripting/actor_func.go)
Returns `true` if actor can be tamed.

## [ActorObject.TrainSkill(skillName string, skillLevel int)](/scripting/actor_func.go)
Sets an ActorObject skill level, if it's greater than what they already have

|  Argument | Explanation |
| --- | --- |
| skillName | The name of the skill to train, such as `map` or `backstab`. |

## [ActorObject.GetSkillLevel(skillName string)](/scripting/actor_func.go)
Returns the current skil level for the skillName, or zero if none.

|  Argument | Explanation |
| --- | --- |
| skillName | The name of the skill to train, such as `map` or `backstab`. |


## [ActorObject.MoveRoom(destRoomId int [, leaveCharmedMobsBehind bool] )](/scripting/actor_func.go)
Quietly moves an ActorObject to a new room

|  Argument | Explanation |
| --- | --- |
| destRoomId | The room id to move them to. |
| leaveCharmedMobsBehind | If true, does not also move charmed mobs with the user. |

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

## [ActorObject.RemoveBuff(buffId int)](/scripting/actor_func.go)
Remove a buff without triggering onEnd

|  Argument | Explanation |
| --- | --- |
| buffId | The ID of the buff to remove |

## [ActorObject.HasItemId(itemId int, [excludeWorn bool]) bool](/scripting/actor_func.go)
Check whether an ActorObject has an item id in their backpack

|  Argument | Explanation |
| --- | --- |
| itemId | The ItemId to check for. |
| itemId (optional) | Ignore worn items? |

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

## [ActorObject.HasSpell(spellId string)](/scripting/actor_func.go)
Returns true if the actor has the spell supplied

|  Argument | Explanation |
| --- | --- |
| spellId | The ID of the spell |

## [ActorObject.LearnSpell(spellId string) bool](/scripting/actor_func.go)
Adds the spell to the Actors spellbook. Returns true if learned, false if already known.

|  Argument | Explanation |
| --- | --- |
| spellId | The ID of the spell |

## [ActorObject.IsAggro(targetActor ActorObject)](/scripting/actor_func.go)
Returns true if the actor is aggro vs targetActor

|  Argument | Explanation |
| --- | --- |
| targetActor | [ActorObject](FUNCTIONS_ACTORS.md) |

## [ActorObject.GetMobKills(mobId int) int](/scripting/actor_func.go)
Returns the number of times the actor has killed a certain mobId

|  Argument | Explanation |
| --- | --- |
| mobId | ID of the mob to check |

## [ActorObject.GetRaceKills(raceName string) int](/scripting/actor_func.go)
Returns the number of times the actor has killed a certain race of mob

|  Argument | Explanation |
| --- | --- |
| raceName | race name such as human, goblin, rodent |

## [ActorObject.GetHealth() int](/scripting/actor_func.go)
Returns current actor health

## [ActorObject.GetHealthMax() int](/scripting/actor_func.go)
Returns current actor max health

## [ActorObject.GetHealthPct() float](/scripting/actor_func.go)
Returns current actor health as a percentage

## [ActorObject.GetMana() int](/scripting/actor_func.go)
Returns current actor mana

## [ActorObject.GetManaMax() int](/scripting/actor_func.go)
Returns current actor max mana

## [ActorObject.GetManaPct() float](/scripting/actor_func.go)
Returns current actor mana as a percentage

## [ActorObject.SetAdjective(adj string, addIt bool)](/scripting/actor_func.go)
Adds or removes a specific text adjective to the characters name

|  Argument | Explanation |
| --- | --- |
| adj | Adjective such as "sleeping", "crying" or "busy" |
| addIt | `true` to add it. `false` to remove it. |

## [ActorObject.IsCharmed( [userId1, userId2, etc ] ) bool ](/scripting/actor_func.go)
Sets a mob to charmed by a user for a set number of rounds.

|  Argument | Explanation |
| --- | --- |
| userId | One or more users to test against. If ommitted, returns true if charmed at all by anyone. |

## [ActorObject.CharmSet(userId int, charmRounds int, [ onRevertCommand1, onRevertCommand2, etc ])](/scripting/actor_func.go)
Sets a mob to charmed by a user for a set number of rounds.

|  Argument | Explanation |
| --- | --- |
| userId | userId that the mob will be charmed to |
| charmRounds | How many rounds it should last, or -2 for unlimited. |
| onRevertCommand | One or more commands for the mob to execute when the charm expires |

## [ActorObject.CharmRemove()](/scripting/actor_func.go)
Immediately discards any charm effect without expiration effects.

## [ActorObject.CharmExpire()](/scripting/actor_func.go)
Forces the current charm of the mob to expire

## [ActorObject.GetCharmCount() int](/scripting/actor_func.go)
Returns the number of charmed creatures in the actors control
