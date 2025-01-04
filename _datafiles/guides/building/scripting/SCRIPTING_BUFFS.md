# Buff Scripting

Example Script: 
* [Buff Definition](/_datafiles/world/default/buffs/1-illumination.yaml)
* [Buff Script](/_datafiles/world/default/buffs/1-illumination.js)

## Script paths

All mob scripts reside in a subfolder of their zone/definition file.

For example, the mob located at  [/_datafiles/world/default/buffs/1-illumination.yaml](/_datafiles/world/default/buffs/1-illumination.yaml) would place its script at [/_datafiles/world/default/buffs/1-illumination.js](/_datafiles/world/default/buffs/1-illumination.js)

# Script Functions and Rules

Mob scripts can maintain their own internal state. If you define or alter a global varaible it will persist until the mob despawns.

The following functions are special keywords that will be invoked under specific circumstances if they are defined within your script:

---

```
function onStart(actor, triggersLeft) {
}
```

`onStart()` is called when a buff is first added to an actor.

|  Argument | Explanation |
| --- | --- |
| actor | [ActorObject](FUNCTIONS_ACTORS.md) |
| triggersLeft | `int` number of triggers until it expires |

---

```
function onTrigger(actor, triggersLeft) {
}
```

`onTrigger()` is called every time a buff triggers.

|  Argument | Explanation |
| --- | --- |
| actor | [ActorObject](FUNCTIONS_ACTORS.md) |
| triggersLeft | `int` number of triggers until it expires |

---

```
function onTrigger(actor, triggersLeft) {
}
```

`onEnd()` is called when a buff has run its course, right before it is removed.

|  Argument | Explanation |
| --- | --- |
| actor | [ActorObject](FUNCTIONS_ACTORS.md) |
| triggersLeft | `int` number of triggers until it expires |

---
