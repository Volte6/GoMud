# Conversations

NPC's (Mobs) can strike up random conversations with eachother.

These conversations are defined in flat files. Each base mob type has a flat file named by the `MobId` and in a folder according to the `Zone` it will be found in.

Example: `Rat` (`MobId 1`): _\_datafiles/conversations/_**frostfang/1.yaml** [link](frostfang/1.yaml)

These `yaml` files are defined as arrays of possible conversations, each defining who can participate and what the actions taken by each participant will be.

* While most "conversations" will usually be speaking, they can do other commands such as `emote`, `attack` and so on. Anything in the mob command set.
* Mobs will only enter conversations randomly if idle. 
* Mobs in conversations will still respond to scripted interactions.
* If mobs are in combat they will not perform their conversational actions.
* Mobs engaged in conversation will do NO idle actions until the conversation is complete. This may mean the second mob may just sit there doing nothing if the conversation is one sided. This may be useful sometimes, such as having a mob leave the room and re-enter one or two rounds later.

# Format

```
- 
  Supported: # A map of lowercase names of "Initiator" (#1) to array of "Participant" (#2) names allowed to use this conversation. 
    "rat": ["rat", "big rat"]
  Conversation:
    - ["#1 sayto #2 SQUEEK!"]
    - ["#2 sayto #1 SQUEEEEEEEK!"]
```

**Supported** - This defines who can initate the conversation, and all mobs names that can be the participant in the conversation. It is a map of `Mob Name` to `Array of Mob Names`. In this manner, when a base MobId is used, but the mob has a different name they can be part of conversations or have their own uniquely defined conversations. 

**Conversation** - This is an `Array of String Arrays`, defining all actions performed, one round at a time. If you want multiple actions to be performed in one round, include them in the same string array. Conversation continues down the Array, one round at a time, until all items have been completed (or otherwise interrupted).

All conversations strings must begin with `#1` or `#2`, indicating which mob in the conversation will perform the action.

Additionally, anywhere else in the string `#1` or `#2` is specified will be replaced with a shorthand identifier of the mob, for use in targetted actions. This will NOT be the name of the mob, so don't use it as such. For example, to have the first mob attack the second would be: `"#1 attack #2"`.

Conversation files can hold multiple conversation entries as shown here:

```
- 
  Supported:
    "rat": ["rat", "big rat"]
  Conversation:
    - ["#1 sayto #2 SQUEEK!"]
    - ["#2 sayto #1 SQUEEEEEEEK!"]
- 
  Supported:
    "rat": ["guard"]
  Conversation:
    - ["#1 sayto #2 SQUEEK!"]
    - ["#2 say Rats! I hate them!",
       "#2 attack #1"]
```
