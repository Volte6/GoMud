# Room Specific Functions

- [Room Specific Functions](#room-specific-functions)
  - [GetRoom(roomId int) RoomObject ](#getroomroomid-int-roomobject-)
  - [RoomObject.RoomId() int](#roomobjectroomid-int)
  - [RoomObject.SendText(msg string\[, excludeUserIds int\])](#roomobjectsendtextmsg-string-excludeuserids-int)
  - [RoomObject.SetTempData(key string, value any)](#roomobjectsettempdatakey-string-value-any)
  - [RoomObject.GetTempData(key string) any](#roomobjectgettempdatakey-string-any)
  - [RoomObject.SetPermData(key string, value any)](#roomobjectsetpermdatakey-string-value-any)
  - [RoomObject.GetPermData(key string) any](#roomobjectgetpermdatakey-string-any)
  - [RoomObject.GetItems() \[\]ItemObject](#roomobjectgetitems-itemobject)
  - [RoomObject.DestroyItem(itm ScriptItem) ](#roomobjectdestroyitemitm-scriptitem-)
  - [RoomObject.SpawnItem(itemId int, inStash bool) \[\]ItemObject](#roomobjectspawnitemitemid-int-instash-bool-itemobject)
  - [RoomObject.GetMobs() \[\]int](#roomobjectgetmobs-int)
  - [RoomObject.GetPlayers() \[\]int](#roomobjectgetplayers-int)
  - [RoomObject.GetContainers() \[\]string](#roomobjectgetcontainers-string)
  - [RoomObject.GetExits() \[\]object](#roomobjectgetexits-object)
  - [GetMap(mapRoomId int, mapSize string, mapHeight int, mapWidth int, mapName string, showSecrets bool \[,mapMarker string, mapMarker string\]) string](#getmapmaproomid-int-mapsize-string-mapheight-int-mapwidth-int-mapname-string-showsecrets-bool-mapmarker-string-mapmarker-string-string)
  - [RoomObject.HasQuest(questId string \[,partyUserId int\]) \[\]int](#roomobjecthasquestquestid-string-partyuserid-int-int)
  - [RoomObject.MissingQuest(questId string \[,partyUserId int\]) \[\]int](#roomobjectmissingquestquestid-string-partyuserid-int-int)
  - [RoomObject.SpawnMob(mobId int) Actor](#roomobjectspawnmobmobid-int-actor)
  - [RoomObject.AddTemporaryExit(exitNameSimple string, exitNameFancy string, exitRoomId int, expiresTimeString string](#roomobjectaddtemporaryexitexitnamesimple-string-exitnamefancy-string-exitroomid-int-expirestimestring-string)
  - [RoomObject.RemoveTemporaryExit(exitNameSimple string, exitNameFancy string, exitRoomId int](#roomobjectremovetemporaryexitexitnamesimple-string-exitnamefancy-string-exitroomid-int)
  - [RoomObject.HasMutator(mutName string) bool](#roomobjecthasmutatormutname-string-bool)
  - [RoomObject.AddMutator(mutName string)](#roomobjectaddmutatormutname-string)
  - [RoomObject.RemoveMutator(mutName string)](#roomobjectremovemutatormutname-string)
  - [RoomObject.RepeatSpawnItem(itemId int, roundInterval int \[, containerName\]](#roomobjectrepeatspawnitemitemid-int-roundinterval-int--containername)
  - [RoomObject.SetLocked(exitName string, lockIt bool)](#roomobjectsetlockedexitname-string-lockit-bool)

## [GetRoom(roomId int) RoomObject ](/internal/scripting/room_func.go)
Retrieves a RoomObject for a given roomId.

## [RoomObject.RoomId() int](/internal/scripting/room_func.go)
Returns the roomId of the room.

## [RoomObject.SendText(msg string[, excludeUserIds int])](/internal/scripting/room_func.go)
Sends a message to everyone in the room.

|  Argument | Explanation |
| --- | --- |
| msg | the message to send |
| excludeUserIds | One or more comma separated userIds to exclude from receiving the message. |

## [RoomObject.SetTempData(key string, value any)](/internal/scripting/room_func.go)
Sets temporary data for the room (Lasts until the room is unloaded from memory).

_Note: This is useful for short term saving/retrieving data between room scripts, such as a switch being triggered._

|  Argument | Explanation |
| --- | --- |
| key | A unique identifier for the data. |
| value | What you will be saving. |

## [RoomObject.GetTempData(key string) any](/internal/scripting/room_func.go)
Gets temporarily saved data for the room. Data is ephemeral.

_Note: This is useful for short term saving/retrieving data between room scripts, such as a switch being triggered._

|  Argument | Explanation |
| --- | --- |
| key | A unique identifier for the data. |

## [RoomObject.SetPermData(key string, value any)](/internal/scripting/room_func.go)
Sets permanent data for the room (Saved even when room is unloaded from memory).

_Note: This is useful for long term saving/retrieving data between room scripts, such as a leaderboard or clan ownership._

|  Argument | Explanation |
| --- | --- |
| key | A unique identifier for the data. |
| value | What you will be saving. |

## [RoomObject.GetPermData(key string) any](/internal/scripting/room_func.go)
Gets permanently saved data for the room.

_Note: This is useful for long term saving/retrieving data between room scripts, such as a leaderboard or clan ownership._

|  Argument | Explanation |
| --- | --- |
| key | A unique identifier for the data. |

## [RoomObject.GetItems() []ItemObject](/internal/scripting/room_func.go)
Returns an array of items on the floor of the room.

_Note: See [/scripting/docs/FUNCTIONS_ITEMS.md](/internal/scripting/docs/FUNCTIONS_ITEMS.md) for details on ItemObject objects._

## [RoomObject.DestroyItem(itm ScriptItem) ](/internal/scripting/room_func.go)
Destroy an item from the ground.

## [RoomObject.SpawnItem(itemId int, inStash bool) []ItemObject](/internal/scripting/room_func.go)
Spawns an item in the room.

|  Argument | Explanation |
| --- | --- |
| itemId | ItemId to spawn. |
| inStash | If true, spawns stashed instead of visible. |

## [RoomObject.GetMobs() []int](/internal/scripting/room_func.go)
Returns an array of `mobInstanceIds` in the room.

## [RoomObject.GetPlayers() []int](/internal/scripting/room_func.go)
Returns an array of `userIds` in the room.

## [RoomObject.GetContainers() []string](/internal/scripting/room_func.go)
Gets a list of container names in the room.

## [RoomObject.GetExits() []object](/internal/scripting/room_func.go)
Gets a list of exits in the room.

|  Argument | Explanation |
| --- | --- |
| roomId | The room id to get containers for. |

Each `object` in the returned array has the following properties:
|  Property | Explanation |
| --- | --- |
| Name | Name of the exit such as `north` or `cave`. |
| Secret | Whether or not the exit is secret/hidden. |
| Lock | `false` if no lock |
| Lock.LockId | Id if the lock (Some keys may match it) |
| Lock.Difficulty | Difficulty rating of the lock |
| Lock.Sequence | Lockpicking sequence of the lock such as `UUDU` |

## [GetMap(mapRoomId int, mapSize string, mapHeight int, mapWidth int, mapName string, showSecrets bool [,mapMarker string, mapMarker string]) string](/internal/scripting/room_func.go)
Gets a rendered map of an area.

|  Argument | Explanation |
| --- | --- |
| mapRoomId | The room id center the map on. |
| mapSize | `wide` or `normal`. Wide maps fit more rooms but don't show the connections. |
| mapHeight | How many lines high the map should be |
| mapWidth | How many lines wide the map should be |
| mapName | The title to display at the top of the map |
| showSecrets | If `true`, show secret rooms. |
| mapMarker (optional) | Zero or more special strings specifying a symbol and legend to override on the map. |
|   | For example: `1,×,Here` Would put `×` on `RoomId 1` and mark is as `Here` on the legend. |

## [RoomObject.HasQuest(questId string [,partyUserId int]) []int](/internal/scripting/room_func.go)
Returns an array of userId's in the room who have the questId. If partyyUserId is supplied, only checks the user and their party specified.

_Note: This could be useful for situations where you want to allow a whole party access to an area even if only one of them has the quest._

|  Argument | Explanation |
| --- | --- |
| questId | The identifier of the quest such as `3-start`. |
| partyUserId (optional) | Only check the specified user and their party |

## [RoomObject.MissingQuest(questId string [,partyUserId int]) []int](/internal/scripting/room_func.go)
Returns an array of userId's in the romo who DON'T have the questId. If partyyUserId is supplied, only checks the user and their party specified.

_Note: This could be useful for situations where you want to disallow a whole party access to an area even if only one of them is missing the quest._

|  Argument | Explanation |
| --- | --- |
| questId | The identifier of the quest such as `3-start`. |
| partyUserId (optional) | Only check the specified user and their party |

## [RoomObject.SpawnMob(mobId int) Actor](/internal/scripting/room_func.go)
Creates a new instance of MobId,and returns the `Actor` of the mob.

|  Argument | Explanation |
| --- | --- |
| mobId | The ID if the mob type to spawn. NOT THE INSTANCE ID. |

## [RoomObject.AddTemporaryExit(exitNameSimple string, exitNameFancy string, exitRoomId int, expiresTimeString string](/internal/scripting/room_func.go)
Adds a temporary exit to the room for the specified amount of time.

|  Argument | Explanation |
| --- | --- |
| exitNameSimple | The simple plain text exit name. |
| exitNameFancy | Should be the simple name, but can have color tags. |
| exitRoomId | The roomId the exit should lead to. |
| expiresTimeString | Time string (1 day, 1 real day, 4 hours, etc) before it vanishes. |

## [RoomObject.RemoveTemporaryExit(exitNameSimple string, exitNameFancy string, exitRoomId int](/internal/scripting/room_func.go)
Removes a temporary exit

_Note: all 3 parameters much match an existing temporary exit for it to be removed._

|  Argument | Explanation |
| --- | --- |
| exitNameSimple | The simple plain text exit name. |
| exitNameFancy | Should be the simple name, but can have color tags. |
| exitRoomId | The roomId the exit should lead to. |


## [RoomObject.HasMutator(mutName string) bool](/internal/scripting/room_func.go)
Returns true if the room has the specified mutator

|  Argument | Explanation |
| --- | --- |
| mutName | the MutatorId of the mutator. |

## [RoomObject.AddMutator(mutName string)](/internal/scripting/room_func.go)
Adds a new mutator to a room.

_Note: If the mutator already exists this is ignored._

|  Argument | Explanation |
| --- | --- |
| mutName | the MutatorId of the mutator. |

## [RoomObject.RemoveMutator(mutName string)](/internal/scripting/room_func.go)
Removes a mutator from a room.

_Note: This only expires it. It may be a mutator that respawns, in which case this doens't really completely remove it._

|  Argument | Explanation |
| --- | --- |
| mutName | the MutatorId of the mutator. |


## [RoomObject.RepeatSpawnItem(itemId int, roundInterval int [, containerName]](/internal/scripting/room_func.go)
Removes a temporary exit

_Note: all 3 parameters much match an existing temporary exit for it to be removed._

|  Argument | Explanation |
| --- | --- |
| itemId | What item? |
| roundInterval | How many rounds until the item respawns after it is taken/removed from the room? |
| containerName | Optional container for the item to spawn into. |

## [RoomObject.SetLocked(exitName string, lockIt bool)](/internal/scripting/room_func.go)
Sets an exit to locked or not (If it has a lock)

|  Argument | Explanation |
| --- | --- |
| exitName | The exitname to lock/unlock |
| lockIt | if true, sets it to locked. Otherwise, unlocks it. |

