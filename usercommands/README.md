# usercommands

The `usecommands` package defines a function type and contains all of the commands a user can enter.

```
type UserCommand func(rest string, userId int) (bool, string, error)
```

All commands follow that definition, where `rest` contains everything except the initial command the user entered, `userId` is who is executing the command, and `cmdQueue` is a way to queue up commands on behalf of users or mobs.

All new commands must return a `util.MessageQueue` containing any messages to go out to rooms/players, an optional error for logging any questionable data:

```

func Glarble(rest string, userId int) (bool, string, error) {
    
    room.SendText(`This glarble goes out to all players in room`)
    user.SendText(`This glarble goes out to the user`)
    response.Handled = true

    return response, nil
}
```

Then the command must be entered into the `userCommands` map in [usercommands.go](/scripting/usercommands.go)


TODO: More info.

