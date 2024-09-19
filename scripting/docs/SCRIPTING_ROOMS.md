# Room Scripting

Example Script: 
* [Room Script](../../_datafiles/rooms/frostfang/1.js)

## Script paths

All room scripts reside the same folder as the room definition file.

For example, the room located at `../../_datafiles/rooms/frostfang/1.yaml` would place its script at `../../_datafiles/rooms/frostfang/1.js`

# Script Functions and Rules

Room scripts can maintain their own internal state. If you define or alter a global varaible it will persist until the room is unloaded from memory.

The following functions are special keywords that will be invoked under specific circumstances if they are defined within your script:

---

```
function onLoad(room RoomObject) {

}
```

`onLoad()` is useful for initializing any state for the room, especially if it might take some extra time. onLoad() is usually given more time to execute than any other function.
It is usually called the first time a player enters a room.

|  Argument | Explanation |
| --- | --- |
| room | [RoomObject](FUNCTIONS_ROOMS.md) |

---

```
function onEnter(user ActorObject, room RoomObject) {
}
```

`onEnter()` is called when a player enters the room.

|  Argument | Explanation |
| --- | --- |
| user | [ActorObject](FUNCTIONS_ACTORS.md) |
| room | [RoomObject](FUNCTIONS_ROOMS.md) |

---

```
function onExit(user ActorObject, room RoomObject) {
}
```

`onExit()` is called when a player exits the room.

|  Argument | Explanation |
| --- | --- |
| user | [ActorObject](FUNCTIONS_ACTORS.md) |
| room | [RoomObject](FUNCTIONS_ROOMS.md) |

---

```
function onCommand(cmd string, rest string, user ActorObject, room RoomObject) {
}
```

`onCommand()` is called if anyone in the room types anything at all (even invalid commands).

Returning `true` will halt any further processing of the response (i.e. "I've handled it"), and returning `false` will all the command to continue along and be processed as normal.

|  Argument | Explanation |
| --- | --- |
| cmd | the command entered, such as `look`, `drop` or `west`. |
| rest | Everything entered after the command (if anything). |
| user | [ActorObject](FUNCTIONS_ACTORS.md) |
| room | [RoomObject](FUNCTIONS_ROOMS.md) |

---

```
function onCommand_{command}(rest string, user ActorObject, room RoomObject) {
}
```

`onCommand_{command}()` is called if anyone in the room types whatever is after the underscore.

For example, `onCommand_look()` would be called if anyone types `look`.

If an `onCommand_{command}` is defined in a script, that command will not be passed to the normal `onCommand()`. So `onCommand_look()` would be called, but `onCommand()` with `look` as a `cmd` value would not.

In all other ways, this follows the same rules as the normal `onCommand()` function.

|  Argument | Explanation |
| --- | --- |
| rest | Everything entered after the command (if anything). |
| user | [ActorObject](FUNCTIONS_ACTORS.md) |
| room | [RoomObject](FUNCTIONS_ROOMS.md) |

---