# Room Specific Functions

---
[RoomSetTempData(roomId int, key string, value any)](room_func.go) - _Sets temporary data for the room (Lasts until the room is unloaded from memory)._

Note: This is useful for saving/retrieving data between room scripts.

|  Argument | Explanation |
| --- | --- |
| roomId | The target room to save data to. |
| key | A unique identifier for the data. |
| value | What you will be saving. |

---

[RoomGetTempData(roomId int, key string) any](room_func.go) - _Gets temporary data for the room._

Note: This is useful for saving/retrieving data between room scripts.

|  Argument | Explanation |
| --- | --- |
| roomId | The target room to retrieve data from. |
| key | A unique identifier for the data. |

---

[RoomGetItems(roomId int) []Item](room_func.go) - _Returns an array of items on the floor of the room._

Note: See [ITEMS.md](ITEMS.md) for details on Item objects.

|  Argument | Explanation |
| --- | --- |
| roomId | The room id to get items for. |

---

[RoomGetMobs(roomId int) []int](room_func.go) - _Returns an array of `mobInstanceIds` in the room._

|  Argument | Explanation |
| --- | --- |
| roomId | The room id to get mobs for. |

---

[RoomGetPlayers(roomId int) []int](room_func.go) - _Returns an array of `userIds` in the room._

|  Argument | Explanation |
| --- | --- |
| roomId | The room id to get userIds for. |

---

[RoomGetContainers(roomId int) []string](room_func.go) - _Gets a list of container names in the room._

|  Argument | Explanation |
| --- | --- |
| roomId | The room id to get containers for. |

---

[RoomGetExits(roomId int) []object](room_func.go) - _Gets a list of container names in the room._

|  Argument | Explanation |
| --- | --- |
| roomId | The room id to get containers for. |

Each `object` in the returned array has the following properties:
|  Property | Explanation |
| --- | --- |
| Name | Name of the exit such as `north` or `cave`. |
| RoomId | The roomId the exit leads to. |
| Secret | Whether or not the exit is secret/hidden. |
| Lock | `false` if no lock |
| Lock.LockId | Id if the lock (Some keys may match it) |
| Lock.Difficulty | Difficulty rating of the lock |
| Lock.Sequence | Lockpicking sequence of the lock such as `UUDU` |

---

[RoomGetMap(mapRoomId int, mapSize string, mapHeight int, mapWidth int, mapName string, showSecrets bool, [,mapMarker string, mapMarker string]) string](room_func.go) - _Gets a rendered map of an area._

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
---

