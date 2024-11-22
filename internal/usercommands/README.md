# usercommands

The `usercommands` package defines a function type and contains all of the commands a user can enter.

```
type UserCommand func(rest string, user *users.UserRecord, room *rooms.Room) (bool, error)
```

All commands follow that definition, where
* `rest` contains everything except the initial command the user entered (if user entered `glarble some text`, `rest` would contain `some text`)
* `user` is the user object executing the command
* `room` is the room object for the room the user is in while executing the command

```

func Glarble(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {
    
    room.SendText(`This glarble text goes out to all players in room`)

    room.SendText(`This glarble text goes out to all players in room except the user who typed the command`, user.UserId)

    user.SendText(`This glarble text goes out to the user`)

    // we handled this command in this function, so return true
    return true, nil
}
```

Then the command must be entered into the `userCommands` map in [usercommands.go](/internal/scripting/usercommands.go)

Each entry into `userCommands` defines:
* `UserCommand` - the function for the command
* `AllowedWhenDowned` - can users execute this command when downed?
* `AdminOnly` - is this a command only admins can use?

Must return two values:
* bool - whether this was handled or should be allowed to continue down processing chain (and indicate a failed command if never gets handled)
* error - An error with error message (for logging purposes)
  
TODO: More info.

