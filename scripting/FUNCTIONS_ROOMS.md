# Room Specific Functions

---

[GetRoom(roomId int) ROOMOBJ ](room_func.go) - _Retrieves a ROOMOBJ for a given roomId._

---

[ROOMOBJ.RoomId() int](room_func.go) - _Returns the roomId of the room._

---

[ROOMOBJ.SetTempData(key string, value any)](room_func.go) - _Sets temporary data for the room (Lasts until the room is unloaded from memory)._

Note: This is useful for saving/retrieving data between room scripts.

|  Argument | Explanation |
| --- | --- |
| key | A unique identifier for the data. |
| value | What you will be saving. |

---

[ROOMOBJ.GetTempData(key string) any](room_func.go) - _Gets temporary data for the room._

Note: This is useful for saving/retrieving data between room scripts.

|  Argument | Explanation |
| --- | --- |
| key | A unique identifier for the data. |

---

[ROOMOBJ.GetItems() []Item](room_func.go) - _Returns an array of items on the floor of the room._

Note: See [ITEMS.md](ITEMS.md) for details on Item objects.

---

[ROOMOBJ.GetMobs() []int](room_func.go) - _Returns an array of `mobInstanceIds` in the room._

---

[ROOMOBJ.GetPlayers() []int](room_func.go) - _Returns an array of `userIds` in the room._

---

[ROOMOBJ.GetContainers() []string](room_func.go) - _Gets a list of container names in the room._

---

[ROOMOBJ.GetExits() []object](room_func.go) - _Gets a list of exits in the room._

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

---

[GetMap(mapRoomId int, mapSize string, mapHeight int, mapWidth int, mapName string, showSecrets bool, [,mapMarker string, mapMarker string]) string](room_func.go) - _Gets a rendered map of an area._

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
