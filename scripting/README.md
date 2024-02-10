# Scripting Language

All scripting is in ECMAScript 5.1 (AKA javascript).

# Room Scripting
See [Room Scripting](ROOM_SCRIPTING.md)

# Mob Scripting
See [Mob Scripting](MOB_SCRIPTING.md)

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

# Script Functions

[Mob Functions](FUNCTIONS_MOBS.md) - Functions that query or alter mob data.

[Room Functions](FUNCTIONS_ROOMS.md) - Functions that query or alter room data.

[User Functions](FUNCTIONS_USERS.md) - Functions that query or alter user data.

[Utility Functions](FUNCTIONS_UTIL.md) - Helper and info functions.

[Messaging Functions](FUNCTIONS_MESSAGING.md) - Helper and info functions.

