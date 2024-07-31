# Item definition

TODO. For now, some examples.

Most important is keeping track of possible statmods.


## Items with stat mods

```
itemid: 20037
name: priests ring
namesimple: ring
description: A hand crafted ring with a priestly insignia.
type: ring
subtype: wearable
damagereduction: 1
statmods:
  #
  # Magic
  #
  casting: 1             # Increases the percentage success for all spell casting
  casting-restoration: 1 # Increases the percentage success on a per-magic-school basis
  #
  # Base stats
  #
  strength: 1            # Increase main strength stat
  speed: 1               # Increase main speed stat
  smarts: 1              # Increase main smarts stat
  vitality: 1            # Increase main vitality stat
  mysticism: 1           # Increase main mysticism stat
  perception: 1          # Increase main perception stat
  #
  # Health/MP
  #
  healthmax: 1           # Increase max health
  manamax: 1             # Increase max mana
  healthrecovery: 1      # Increase the health you recover every "recover" event
  manarecovery: 1        # Increase the mana you recover every "recover" event
```


## Keys

```
itemid: 3
name: crypt key
namesimple: key
description: An ancient and crude key. I wonder where it could be used?
type: key
subtype: generic
keylockid: 110-west
```

## Items with uses

```
itemid: 22
name: training coupon
namesimple: coupon
displayname: ":coupon"
description: Grants a training point when used.
type: object
subtype: usable
uses: 1
value: 1
```


## Items with buffs on use

```
itemid: 30012
name: phoenix elixir
namesimple: tea
description: A fiery potion made from sunfire blooms that revitalizes and boosts energy levels.
type: drink
subtype: drinkable
uses: 1
value: 92
buffids: 
- 14
- 16
```

# Items with buffs on wear

_Note: These buffs remain in play while worn, and should not expire on a timer_

```
itemid: 10006
name: glowing battleaxe
displayname: ":glowing"
namesimple: battleaxe
description: The battleaxe, robust and formidable, commands attention with its radiant glow that emanates from the core of its blade.
type: weapon
hands: 2
subtype: cleaving
damage:
  diceroll: 2d10+1
statmods:
  speed: -10
cursed: true
wornbuffids:
  - 1
```