
// This mob is in 271.yaml
/*
  idlecommands:
  - ifnotquest 4-end say I'm so hungry.
  - ifnotquest 4-end emote pats his belly as it grumbles.
  - wander
  itemtrades:
  - accepteditemids: [30004]
    prizequestids: [4-end]
    prizecommands: [say Thanks! I can get on with my day now.]
  asksubjects:
  - ifquest: ""
    ifnotquest: 4-start
    asknouns:
    - quest
    - hunger
    - hungry
    - belly
    - food
    replycommands:
    - say I forgot my lunch today, and I'm so hungry.
    - say Do you think you could find a cheese sandwich for me?
    - givequest 4-start
  scripttag: hungry

*/

const nouns = ["quest", "hunger", "hungry", "belly", "food"]

// eventDetails.sourceId   - mobInstanceId or userId
// eventDetails.sourceType - mob or user
function onCommand_wave(rest, mobInstanceId, roomId, eventDetails) {
    MobCommand(mobInstanceId, "wave")
}

// eventDetails.sourceId   - mobInstanceId or userId
// eventDetails.sourceType - mob or user
function onCommand(cmd, rest, mobInstanceId, roomId, eventDetails) {

    if (cmd == "wave") {
        MobCommand(mobInstanceId, "wave")
    }
    return false;
}

// Invoked when asked a question
// Intended for quests, but can be used for other things
// eventDetails.askText    - Text asked by the user
// eventDetails.sourceId   - mobInstanceId or userId
// eventDetails.sourceType - mob or user
function onAsk(mobInstanceId, roomId, eventDetails) {

    parts = eventDetails.askText.toLowerCase().split(' ');
    for (var i = 0; i < parts.length; i++) {
        match = UtilFindMatchIn(parts[i], nouns);
        if ( match.exact.length > 0 ) {

            MobCommand(mobInstanceId, "emote rubs his belly.")
            MobCommand(mobInstanceId, "say I forgot my lunch today, and I'm so hungry.")
            MobCommand(mobInstanceId, "say Do you think you could find a cheese sandwich for me?")

            UserGiveQuest(eventDetails.sourceId,  "4-start")

            return true;
        }
    }

    return false;
}

// Invoked when given an item
// eventDetails.sourceId   - mobInstanceId or userId
// eventDetails.sourceType - mob or user
// eventDetails.gold       - 0+
// eventDetails.item       - items.Item
function onGive(mobInstanceId, roomId, eventDetails) {

    if (eventDetails.sourceType == "mob") {
        return false;
    }

    if (eventDetails.item.ItemId != 30004) {
        MobCommand(mobInstanceId, "look !"+String(eventDetails.item.ItemId))
        MobCommand(mobInstanceId, "drop !"+String(eventDetails.item.ItemId), UtilGetSecondsToTurns(5))
        return true;
    }

    if ( UserHasQuest(eventDetails.sourceId, "4-start") ) {

        UserGiveQuest(eventDetails.sourceId,  "4-end")
        MobCommand(mobInstanceId, "say Thanks! I can get on with my day now.")
        MobCommand(mobInstanceId, "eat !"+String(eventDetails.item.ItemId), )

        return true;
    }


}

playersTold = {}

// Invoked once every round if mob is idle
function onIdle(mobInstanceId, roomId) {

    round = UtilGetRoundNumber();

    grumbled = false
    userIds = RoomGetPlayers(roomId);

    if ( userIds.length > 0 ) {
        for (var i = 0; i < userIds.length; i++) {

            if ( userIds[i] in playersTold ) {
                if ( round < playersTold[userIds[i]] ) {
                    continue;
                }
            }

            if ( !UserHasQuest(userIds[i], "4-start") ) {
                if ( !grumbled ) {
                    MobCommand(mobInstanceId, "emote pats his belly as it grumbles.");
                    grumbled = true;
                }
                MobCommand(mobInstanceId, "sayto @" + String(userIds[i]) + " I'm so hungry.");
            }

            playersTold[userIds[i]] = round + 5;
        }
        return true;
    }

    for (var key in playersTold) {
        if ( playersTold[key] < round-100 ) {
            delete playersTold[key];
        }
    }

    action = round % 3;

    if ( action == 0 ) {
        MobCommand(mobInstanceId, "wander")
        return true;
    }

    return false;
}

// Invoked when script is first loaded.
// onLoad() is potentially more forgiving of running long
// ScriptLoadTimeoutMs config - so can be used to set up intial state
function onLoad(mobInstanceId) {

}
