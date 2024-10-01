# mobcommands

The `mobcommands` package defines a function type and contains all of the commands a user can enter.

```
type MobCommand func(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error)
```

**Note:** See `usercommands/README.md`

Differences:

* `mobCommand` instead of `userCommands` for registering commands
* `mobCommand` entries do not specify `AdminOnly`
  