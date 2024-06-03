# Room Specific Functions

- [Room Specific Functions](#room-specific-functions)
  - [GetRoom(roomId int) RoomObject ](#getroomroomid-int-roomobject-)
  - [RoomObject.RoomId() int](#roomobjectroomid-int)
  - [RoomObject.SetTempData(key string, value any)](#roomobjectsettempdatakey-string-value-any)
  - [RoomObject.GetTempData(key string) any](#roomobjectgettempdatakey-string-any)
  - [RoomObject.GetItems() \[\]ItemObject](#roomobjectgetitems-itemobject)
  - [RoomObject.GetMobs() \[\]int](#roomobjectgetmobs-int)
  - [RoomObject.GetPlayers() \[\]int](#roomobjectgetplayers-int)
  - [RoomObject.GetContainers() \[\]string](#roomobjectgetcontainers-string)
  - [RoomObject.GetExits() \[\]object](#roomobjectgetexits-object)
  - [GetMap(mapRoomId int, mapSize string, mapHeight int, mapWidth int, mapName string, showSecrets bool \[,mapMarker string, mapMarker string\]) string](#getmapmaproomid-int-mapsize-string-mapheight-int-mapwidth-int-mapname-string-showsecrets-bool-mapmarker-string-mapmarker-string-string)
  - [RoomObject.HasQuest(questId string \[,partyUserId int\]) \[\]int](#roomobjecthasquestquestid-string-partyuserid-int-int)
  - [RoomObject.MissingQuest(questId string \[,partyUserId int\]) \[\]int](#roomobjectmissingquestquestid-string-partyuserid-int-int)
  - [RoomObject.SpawnMob(mobId int) int](#roomobjectspawnmobmobid-int-int)

## [GetRoom(roomId int) RoomObject ](/scripting/room_func.go)
Retrieves a RoomObject for a given roomId.

## [RoomObject.RoomId() int](/scripting/room_func.go)
Returns the roomId of the room.

## [RoomObject.SetTempData(key string, value any)](/scripting/room_func.go)
Sets temporary data for the room (Lasts until the room is unloaded from memory).

_Note: This is useful for saving/retrieving data between room scripts._

|  Argument | Explanation |
| --- | --- |
| key | A unique identifier for the data. |
| value | What you will be saving. |

## [RoomObject.GetTempData(key string) any](/scripting/room_func.go)
Gets temporary data for the room.

_Note: This is useful for saving/retrieving data between room scripts._

|  Argument | Explanation |
| --- | --- |
| key | A unique identifier for the data. |

## [RoomObject.GetItems() []ItemObject](/scripting/room_func.go)
Returns an array of items on the floor of the room.

_Note: See [/scripting/docs/FUNCTIONS_ITEMS.md](/scripting/docs/FUNCTIONS_ITEMS.md) for details on ItemObject objects._

## [RoomObject.GetMobs() []int](/scripting/room_func.go)
Returns an array of `mobInstanceIds` in the room.

## [RoomObject.GetPlayers() []int](/scripting/room_func.go)
Returns an array of `userIds` in the room.

## [RoomObject.GetContainers() []string](/scripting/room_func.go)
Gets a list of container names in the room.

## [RoomObject.GetExits() []object](/scripting/room_func.go)
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

## [GetMap(mapRoomId int, mapSize string, mapHeight int, mapWidth int, mapName string, showSecrets bool [,mapMarker string, mapMarker string]) string](/scripting/room_func.go)
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

## [RoomObject.HasQuest(questId string [,partyUserId int]) []int](/scripting/room_func.go)
Returns an array of userId's in the room who have the questId. If partyyUserId is supplied, only checks the user and their party specified.

_Note: This could be useful for situations where you want to allow a whole party access to an area even if only one of them has the quest._

|  Argument | Explanation |
| --- | --- |
| questId | The identifier of the quest such as `3-start`. |
| partyUserId (optional) | Only check the specified user and their party |

## [RoomObject.MissingQuest(questId string [,partyUserId int]) []int](/scripting/room_func.go)
Returns an array of userId's in the romo who DON'T have the questId. If partyyUserId is supplied, only checks the user and their party specified.

_Note: This could be useful for situations where you want to disallow a whole party access to an area even if only one of them is missing the quest._

|  Argument | Explanation |
| --- | --- |
| questId | The identifier of the quest such as `3-start`. |
| partyUserId (optional) | Only check the specified user and their party |

## [RoomObject.SpawnMob(mobId int) int](/scripting/room_func.go)
Creates a new instance of MobId,and returns the `mobInstanceId` of the mob.

|  Argument | Explanation |
| --- | --- |
| mobId | The ID if the mob type to spawn. NOT THE INSTANCE ID. |
