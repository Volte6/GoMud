
const nouns = ["quest", "hunger", "hungry", "belly", "food"]

// eventDetails.sourceId   - mobInstanceId or userId
// eventDetails.sourceType - mob or user
function onCommand_wave(rest, mob, room, eventDetails) {
    mob.Command("wave")
}

// eventDetails.sourceId   - mobInstanceId or userId
// eventDetails.sourceType - mob or user
function onCommand(cmd, rest, mob, room, eventDetails) {
    if (cmd == "wave") {
        mob.Command("wave")
    }
    return false;
}

// Invoked when asked a question
// Intended for quests, but can be used for other things
// eventDetails.askText    - Text asked by the user
// eventDetails.sourceId   - mobInstanceId or userId
// eventDetails.sourceType - mob or user
function onAsk(mob, room, eventDetails) {

    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }

    parts = eventDetails.askText.toLowerCase().split(' ');
    for (var i = 0; i < parts.length; i++) {
        match = UtilFindMatchIn(parts[i], nouns);
        if ( match.exact.length > 0 ) {

            mob.Command("emote rubs his belly.")
            mob.Command("say I forgot my lunch today, and I'm so hungry.")
            mob.Command("say Do you think you could find a cheese sandwich for me?")

            user.GiveQuest("4-start")

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
function onGive(mob, room, eventDetails) {

    if (eventDetails.sourceType == "mob") {
        return false;
    }

    if (eventDetails.item.ItemId != 30004) {
        mob.Command("look !"+String(eventDetails.item.ItemId))
        mob.Command("drop !"+String(eventDetails.item.ItemId), UtilGetSecondsToTurns(5))
        return true;
    }

    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }

    if ( user.HasQuest("4-start") ) {

        user.GiveQuest("4-end")
        mob.Command("say Thanks! I can get on with my day now.")
        mob.Command("eat !"+String(eventDetails.item.ItemId), )

        return true;
    }


}

playersTold = {}

// Invoked once every round if mob is idle
function onIdle(mob, room) {

    round = UtilGetRoundNumber();

    grumbled = false
    userIds = room.GetPlayers();

    if ( userIds.length > 0 ) {
        for (var i = 0; i < userIds.length; i++) {

            if ( userIds[i] in playersTold ) {
                if ( round < playersTold[userIds[i]] ) {
                    continue;
                }
            }

            if ( (user = GetUser(userIds[i])) == null ) {
                continue;
            }

            if ( !user.HasQuest("4-start") ) {
                if ( !grumbled ) {
                    mob.Command("emote pats his belly as it grumbles.");
                    grumbled = true;
                }
                mob.Command("sayto @" + String(userIds[i]) + " I'm so hungry.");
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
        mob.Command("wander")
        return true;
    }

    return false;
}
