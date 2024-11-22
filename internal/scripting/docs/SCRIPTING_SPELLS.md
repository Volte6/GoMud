# Spell Scripting

Example Script: 
* [Spell Script](../../../_datafiles/spells/heal.js)

## Script paths

All spell scripts reside the same folder as the spell definition file.

For example, the spell located at `../../../_datafiles/spells/heal.yaml` would place its script at `../../../_datafiles/spells/heal.js`

# Script Functions and Rules

The following functions are special keywords that will be invoked under specific circumstances if they are defined within your script:

*IMPORTANT NOTE* - The second argument will vary based on the type of spell:
* neutral - second argument will be a string of what the user typed after `cast <spellname>`.
* harmsingle - Second argument will be a single [ActorObject](FUNCTIONS_ACTORS.md)
* helpsingle - Second argument will be a single [ActorObject](FUNCTIONS_ACTORS.md)
* harmmulti - Second argument will be an array of [ActorObject](FUNCTIONS_ACTORS.md)
* helpmulti - Second argument will be a array of [ActorObject](FUNCTIONS_ACTORS.md)

---

```
function onCast(sourceActor, ARG) {
}
```

`onCast()` is called when a player attempts to cast a spell. Return `false` to ignore the attempt/abort the cast.

|  Argument | Explanation |
| --- | --- |
| sourceActor | [ActorObject](FUNCTIONS_ACTORS.md) |
| ARG | [ActorObject](FUNCTIONS_ACTORS.md) / [[]ActorObject](FUNCTIONS_ACTORS.md) / string |

---

```
function onWait(sourceActor, ARG) {
}
```

`onWait()` is called each round that a player waits for the spell to be officially cast.

|  Argument | Explanation |
| --- | --- |
| sourceActor | [ActorObject](FUNCTIONS_ACTORS.md) |
| ARG | [ActorObject](FUNCTIONS_ACTORS.md) / [[]ActorObject](FUNCTIONS_ACTORS.md) / string |

---

```
function onMagic(sourceActor, ARG) {
}
```

`onMagic()` is called when the spell is ready and finally cast. For harmful type spells, return `true` to abort auto-retaliation from mobs.

|  Argument | Explanation |
| --- | --- |
| sourceActor | [ActorObject](FUNCTIONS_ACTORS.md) |
| ARG | [ActorObject](FUNCTIONS_ACTORS.md) / [[]ActorObject](FUNCTIONS_ACTORS.md) / string |

---


