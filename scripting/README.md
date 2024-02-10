# Scripting

## Room Scripts
TODO: _Information on scripting rooms_

## Mob Scripts
TODO: _Information on scripting mobs_

## Script Functions

### Mobs

---
[MobGetCharacterName(mobInstanceId int) string](scripting/mob_func.go) - _Returns the name of a given mob instance id.
|  Argument | Explanation |
|--- |--- |
| mobInstanceId | The mob this applies to. |
---

[MobCommand(mobInstanceId int, cmd string [,waitTurns int])](scripting/mob_func.go) - _Forces a mob to perform an in-game command._
|  Argument | Explanation |
|--- |--- |
| mobInstanceId | The mob this applies to. |
| cmd | The command to run such as `say hi mom` |
| waitTurns (optional) | Delay this many turns (NOT rounds!) before executing the command. |
---

[MobGetCharacterName(mobInstanceId int) string](scripting/mob_func.go) - _Returns the name of a given mob instance id._
|  Argument | Explanation |
|--- |--- |
| mobInstanceId | The mob this applies to. |
---

[MobCharmed(mobInstanceId int [, userId int]) bool](scripting/mob_func.go) - _Checks whether a mob is charmed by any player._
|  Argument | Explanation |
|--- |--- |
| mobInstanceId | The mob this applies to. |
| userId (optional) | The user to check whether the mob is charmed to. Omit for `any` |
---

[MobCharmSet(mobInstanceId int, userId int, charmRounds int [, onRevertCommand string])](scripting/mob_func.go) - _Sets a mob to charmed._
|  Argument | Explanation |
|--- |--- |
| mobInstanceId | The mob this applies to. |
| userId (optional) | The user who the mob will be charmed to. |
| charmRounds | How many rounds until the charm wears off, or -1 for never. |
| onRevertCommand (optional) | A command the mob will execute when the charm wears off. |
---

[MobCharmRemove(mobInstanceId int)](scripting/mob_func.go) - _Removes a charm without expiring it. If an `onRevertCommand` was supplied with the initial charm, it will not execute._
|  Argument | Explanation |
|--- |--- |
| mobInstanceId | The mob this applies to. |
---

[MobCharmExpire(mobInstanceId int)](scripting/mob_func.go) - _Expires an existing charm. This would still execute any `onRevertCommand`` that was specified with the initial charm._
|  Argument | Explanation |
|--- |--- |
| mobInstanceId | The mob this applies to. |
---

# Special symbols in user or mob commands:

There are some special prefixes that can help target more specifically than just a name.
These are particularly helpful when there may be other matching targets on a given name:
* `goblin` peaceful vs `goblin` that hit you `(that is fighting), 
* `dagger` vs a `dagger` with enhancements
* user `sam` vs user `samuel`, when sam has already left the room.

These are only useful in `commands` such as `look`, `give`, `attack`, etc.

* `!{number}` - denotes a specific `ItemId` as a target. 
  * `drop !123` will drop `ItemId`=`123`
  * `give !123 to samuel` will give `ItemId`=`123` to a user or mob in the room named `samuel`
* `#{number}` - denotes a specific `Mob Instance Id` as a target.
  * `kick #98` will kick `Mob Instance Id`=`98`
* `@{number}` - denotes a specific `UserId` as a target.
  * `give !123 to @5` will give `ItemId`=`123` to `UserId`=`5`
  * `give !123 to #98` will give `ItemId`=`123` to `Mob Instance Id`=`98`

These are optional, everything can still be referred to by `name` or augmented name ( `dagger#2` etc. )
