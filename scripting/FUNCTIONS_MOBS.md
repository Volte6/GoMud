# Mobs Specific Functions

---

[GetMob(mobInstanceId int) MOBOBJ ](mob_func.go) - _Retrieves a MOBOBJ for a given mobInstanceId._

---

[MOBOBJ.InstanceId() int](mob_func.go) - _Returns the mobInstanceId of the MOBOBJ._

---

[MOBOBJ.MobTypeId() int](mob_func.go) - _Returns the base mobId used to spawn new instances._

---
[MOBOBJ.GetCharacterName() string](mob_func.go) - _Returns the name of a given mob._

---

[MOBOBJ.Command(cmd string [,waitTurns int])](mob_func.go) - _Forces a mob to perform an in-game command._

|  Argument | Explanation |
| --- | --- |
| cmd | The command to run such as `say hi mom` |
| waitTurns (optional) | Delay this many turns (NOT rounds!) before executing the command. |

---

[MOBOBJ.IsCharmed([, userId int]) bool](mob_func.go) - _Checks whether a mob is charmed by any player._

|  Argument | Explanation |
| --- | --- |
| userId (optional) | The user to check whether the mob is charmed to. Omit for `any` |

---

[MOBOBJ.CharmSet(userId int, charmRounds int [, onRevertCommand string])](mob_func.go) - _Sets a mob to charmed._

|  Argument | Explanation |
| --- | --- |
| userId (optional) | The user who the mob will be charmed to. |
| charmRounds | How many rounds until the charm wears off, or -1 for never. |
| onRevertCommand (optional) | A command the mob will execute when the charm wears off. |

---

[MOBOBJ.CharmRemove()](mob_func.go) - _Removes a charm without expiring it. If an `onRevertCommand` was supplied with the initial charm, it will not execute._

---

[MOBOBJ.CharmExpire()](mob_func.go) - _Expires an existing charm. This would still execute any `onRevertCommand`` that was specified with the initial charm._

---
