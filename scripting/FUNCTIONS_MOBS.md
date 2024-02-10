# Mobs Specific Functions

---

[MobGetCharacterName(mobInstanceId int) string](mob_func.go) - _Returns the name of a given mob instance id._

|  Argument | Explanation |
| --- | --- |
| mobInstanceId | The mob this applies to. |

---

[MobCommand(mobInstanceId int, cmd string [,waitTurns int])](mob_func.go) - _Forces a mob to perform an in-game command._

|  Argument | Explanation |
| --- | --- |
| mobInstanceId | The mob this applies to. |
| cmd | The command to run such as `say hi mom` |
| waitTurns (optional) | Delay this many turns (NOT rounds!) before executing the command. |

---

[MobIsCharmed(mobInstanceId int [, userId int]) bool](mob_func.go) - _Checks whether a mob is charmed by any player._

|  Argument | Explanation |
| --- | --- |
| mobInstanceId | The mob this applies to. |
| userId (optional) | The user to check whether the mob is charmed to. Omit for `any` |

---

[MobCharmSet(mobInstanceId int, userId int, charmRounds int [, onRevertCommand string])](mob_func.go) - _Sets a mob to charmed._

|  Argument | Explanation |
| --- | --- |
| mobInstanceId | The mob this applies to. |
| userId (optional) | The user who the mob will be charmed to. |
| charmRounds | How many rounds until the charm wears off, or -1 for never. |
| onRevertCommand (optional) | A command the mob will execute when the charm wears off. |

---

[MobCharmRemove(mobInstanceId int)](mob_func.go) - _Removes a charm without expiring it. If an `onRevertCommand` was supplied with the initial charm, it will not execute._

|  Argument | Explanation |
| --- | --- |
| mobInstanceId | The mob this applies to. |

---

[MobCharmExpire(mobInstanceId int)](mob_func.go) - _Expires an existing charm. This would still execute any `onRevertCommand`` that was specified with the initial charm._

|  Argument | Explanation |
| --- | --- |
| mobInstanceId | The mob this applies to. |

---
