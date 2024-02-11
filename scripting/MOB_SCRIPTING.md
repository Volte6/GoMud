# Mob Scripting

Example Script: 
* [Mob Script Tag Instance Script (hungry)](../_datafiles/mobs/frostfang/scripts/2-hungry.js)
* [Mob Script Tag defined in Spawninfo (hungry)](../_datafiles/rooms/frostfang/271.yaml)

## Script paths

All mob scripts reside in a subfolder of their zone/definition file.

For example, the mob located at `_datafiles/mobs/frostfang/2.yaml` would place its script at `_datafiles/mobs/frostfang/scripts/2.js`

If a mob defined in a rooms spawninfo has a `scripttag` defined, it will be appended to the mobs script path with a hyphen. 

For example, `scripttag: hungry` for mob `2` (as above) would load the script `_datafiles/mobs/frostfang/scripts/2-hungry.js`

In this way you can have generic scripts for a mob id, or specific scripts for special rooms or circumstances.

# Script Functions and Rules

Mob scripts can maintain their own internal state. If you define or alter a global varaible it will persist until the mob despawns.

The following functions are special keywords that will be invoked under specific circumstances if they are defined within your script:

---

```
function onLoad(mob MOBOBJ) {

}
```

`onLoad()` is useful for initializing any state for the mob, especially if it might take some extra time. onLoad() is usually given more time to execute than any other function.
It is usually called within one round of a mob instance being created, or possibly sooner if players are present.

|  Argument | Explanation |
| --- | --- |
| mob | [MOBOBJ](FUNCTIONS_MOBS.md) |

---

```
function onIdle(mob MOBOBJ, room ROOMOBJ) {
}
```

`onIdle()` is called each round that a mob isn't in combat or doing something that supercedes being idle, such as trying to walk home after wandering too far.

|  Argument | Explanation |
| --- | --- |
| mob | [MOBOBJ](FUNCTIONS_MOBS.md) |
| room | [ROOMOBJ](FUNCTIONS_ROOMS.md) |

---

```
function onGive(mob MOBOBJ, room ROOMOBJ, eventDetails object) {
}
```

`onGive()` is called when an object or gold is given to a mob. Returning `true` from this function will stop the mob from attempting to wear the object (if applicable).

|  Argument | Explanation |
| --- | --- |
| mob | [MOBOBJ](FUNCTIONS_MOBS.md) |
| room | [ROOMOBJ](FUNCTIONS_ROOMS.md) |
| eventDetails.sourceId | The `userId` or `mobInstanceId` that gave the item/gold |
| eventDetails.sourceType | `"user"` or `"mob"`, the source type of the gift |
| eventDetails.gold | How much gold was given (if any) |
| eventDetails.item | An Item object that was given (if any) |

---

```
function onAsk(mob MOBOBJ, room ROOMOBJ, eventDetails object) {
}
```

`onAsk()` is called when mob is asked something. Returning `false` will cause the mob to respond with a generic rejection such as "the mob just shakes their head".

|  Argument | Explanation |
| --- | --- |
| mob | [MOBOBJ](FUNCTIONS_MOBS.md) |
| room | [ROOMOBJ](FUNCTIONS_ROOMS.md) |
| eventDetails.sourceId | The `userId` or `mobInstanceId` that asked the question |
| eventDetails.sourceType | `"user"` or `"mob"`, the source type of the question |
| eventDetails.askText | The question the mob was asked |

---

```
function onCommand(cmd string, rest string, mob MOBOBJ, room ROOMOBJ, eventDetails object) {
}
```

`onCommand()` is called if anyone in the room types anything at all (even invalid commands).

Returning `true` will halt any further processing of the response, and returning `false` will all the command to continue along and be processed as normal.

NOTE: This is called BEFORE the room's `onCommand()` functions.
NOTE: Other mobs may be in the room with `onCommand()` functions defined, and they may prevent down-stream mobs from being called if they return `true`.

|  Argument | Explanation |
| --- | --- |
| cmd | the command entered, such as `look`, `drop` or `west`. |
| rest | Everything entered after the command (if anything). |
| mob | [MOBOBJ](FUNCTIONS_MOBS.md) |
| room | [ROOMOBJ](FUNCTIONS_ROOMS.md) |
| eventDetails.sourceId | The `userId` or `mobInstanceId` that sent the command |
| eventDetails.sourceType | `"user"` or `"mob"`, the source type of the command |

---

```
function onCommand_{command}(rest, mob MOBOBJ, room ROOMOBJ, eventDetails object) {
}
```

`onCommand_{command}()` is called if anyone in the room types whatever is after the underscore. 

For example, `onCommand_look()` would be called if anyone types `look`.

If an `onCommand_{command}` is defined in a script, that command will not be passed to the normal `onCommand()`. So `onCommand_look()` would be called, but `onCommand()` with `look` as a `cmd` value would not.

In all other ways, this follows the same rules as the normal `onCommand()` function.

|  Argument | Explanation |
| --- | --- |
| rest | Everything entered after the command (if anything). |
| mob | [MOBOBJ](FUNCTIONS_MOBS.md) |
| room | [ROOMOBJ](FUNCTIONS_ROOMS.md) |
| eventDetails.sourceId | The `userId` or `mobInstanceId` that sent the command |
| eventDetails.sourceType | `"user"` or `"mob"`, the source type of the command |

---