# Plugin Support

* Consider changing `usercommands` and `mobcommands` to use a plugin architecture.

# Embedded scripting

* Made an attempt to embed a javascript scripting engine (https://github.com/dop251/goja) but it would hang indefinitely when compiling on a `raspberry pi zero 2 w`, and I removed this for the time being.
* I would like to be able to write generic scripting rules with defined interfaces and exported variables.

# Thoughts on a magic system

* Generic magic as usual might be uninteresting. Consider some other ideas.
* Maybe regularly changing magic phrases that have to be memorized or scribed to a magic book, when chanted they are cast and removed from the book.
* The spells can be rewritten from whatever source, where the magic phrase will have changed? 

# Macros

Implement macros as follows
`=1` - Execute a macro
`=1 look` - Bind a macro to `=1` that will execute `look`
Store macros in UserRecord.Macros
No more than 10 (1-0)

Need to create an InputHandler that will capture the macro and replace it with the intended text.
Alternatively could handle in game world, but may be useful to have this processed before game logic.... Maybe 

# Ascii art

https://www.asciiart.eu/


# Notes on triggers and props

To setup a trigger/prop who's sole purpose is to interrupt a user command and/or block progress conditionally, keep a prop/trigger bare such as:
```
props:
- requiresbuffid: 2
  messagerejectgeneric: The guards step before you. "You must be invited to enter
    the castle," they say. "We cannot let you pass."
  verbs:
  - north
```

The above would interrupt the "north" command, and send back the `messagerejectgeneric`, unless the player has the required buff id

# Concurrent reads/writes

Create more interface methods to RWLock/RWUnlock properties.

# TODO

* Create a skill for using magical objects
  * Possibly not use a charge?

* Potion Mixing?
  * INPUT items
  * OUTPUT a unique potion that provides a buff?

* Mutable buff characteristics?
  * Problematic since buffs are determined by ID currently.

* Shapechanging via race changes? Should work... just need to track reverting back.
  * Could be a buff, would need mutable buffs to know what to turn back into, or track this data somehow.

# Basic thoughts on weapons damage guidelines

Dice qty goes up as weapon gets more special
Bonus damage due to enchantment or special quality ("Fine")

- Sticks/Junks etc.          1d2
- 1H Sm (Daggers, claws)     1d4
  - (+Speed)
- 1H Md 
  - Slashing/Stabbing        1d6
  - Cleaving                 1d8
    - (-Speed)
- 2H Md
  - Slashing/Stabbing        1d8
  - Cleaving                 1d8
    - (-Speed)
- 2H Lg
  - Slashing/Stabbing        1d10
    - (-Speed)
  - Cleaving                 1d10
    - (-Speed)

# Weapon speeds?

Currently weapons have a number of rounds extra they must wait between attacks: `waitrounds: 1` for example.

Consider other ways of scaling weapon speed. For example, in order to achieve an extra attack every 2 or 3 rounds.

This would probably require some sort of "energy" pool that gets incremented by round and decremented by attack.
Weapons would need a "speed" or "energy" characteristic that would drive this (possibly a float scaler?).

Example:
`1.0` - default
`0.5` - 1 attack every 2 rounds
`2.0` - 2 attacks every 1 round

What would this mean for weapons that have extra attacks already? Those would be freebies?

# Notes on creating quests

Players can be given quests through a couple methods:

## Given by mobs:

*Given by a mob idlecommands:*
```
```

*Given by a mob ask response:*
```
 asksubjects:
  - ifquest: ""
    ifnotquest: 1-start
    asknouns:
    - frog
    replycommands:
    - emote sighs.
    - say Please find my frog.
    - givequest 1-start {userid}
```
*Given as a mob trade*
```
  itemtrades:
  - accepteditemids: [20025]
    prizeitemids: [20015]
    prizequestids: [1-end]
```

## Given by room triggers

```
- nouns: [frog]
  verbs: [get]
  trigger:
    questtoken: 2-frogfound
    affected: player
    descriptionplayer: You reach out and snatch up the frog.
    descriptionroom: <ansi fg="username">%s</ansi> snatches up the frog.
```

## Given by picking up /stealing/buying/being given an item

```
itemid: 123123123
name: golden key
namesimple: key
description: A golden key.
questtoken: 56-key
type: key
subtype: generic
```

When putting requirements such as "must have" quest tokens or "must not have" quest tokens, all quests count as "have" if they occured before the current step the player is on.
For example, if a quest has 3 steps: start, search, end
If they are on step "search", and you require they have quest "start" they will, but they will not yet have quest "end"

# Combat ideas

* When anyone in the party is hated/aggrod, entire party becomes a viable target
  * Includ eother players AND MOBS in the potential target list.

* Consider allowing mobs to change targets mid-fight? Should be transparent though... not another "Gets ready to attack" announcement.

# Mob ideas

* Movement along a pre-scripted path?

* Allow mobs to party/group and move together?
