# usercommands

The `usecommands` package defines a function type and contains all of the commands a user can enter.

```
type UserCommand func(rest string, user *users.UserRecord, room *rooms.Room) (bool, error)
```

All commands follow that definition, where `rest` contains everything except the initial command the user entered, `userId` is who is executing the command.

```

func Glarble(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {
    
    room.SendText(`This glarble goes out to all players in room`)
    user.SendText(`This glarble goes out to the user`)

    return response, nil
}
```

Then the command must be entered into the `userCommands` map in [usercommands.go](/scripting/usercommands.go)

Must return two values:
* bool - whether this was handled or should be allowed to continue down processing chain (and indicate a failed command if never gets handled)
* error - An error with error message (for logging purposes)
  
TODO: More info.

