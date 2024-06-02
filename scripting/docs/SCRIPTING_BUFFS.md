# Buff Scripting

Example Script: 
* [Mob Script Tag Instance Script (hungry)](../_datafiles/mobs/frostfang/scripts/2-hungry.js)
* [Mob Script Tag defined in Spawninfo (hungry)](../_datafiles/rooms/frostfang/271.yaml)

## Script paths

All mob scripts reside in a subfolder of their zone/definition file.

For example, the mob located at `../_datafiles/mobs/frostfang/2.yaml` would place its script at `../_datafiles/mobs/frostfang/scripts/2.js`

If a mob defined in a rooms spawninfo has a `scripttag` defined, it will be appended to the mobs script path with a hyphen. 

For example, `scripttag: hungry` for mob `2` (as above) would load the script `../_datafiles/mobs/frostfang/scripts/2-hungry.js`

In this way you can have generic scripts for a mob id, or specific scripts for special rooms or circumstances.

# Script Functions and Rules

Mob scripts can maintain their own internal state. If you define or alter a global varaible it will persist until the mob despawns.

The following functions are special keywords that will be invoked under specific circumstances if they are defined within your script:

---
