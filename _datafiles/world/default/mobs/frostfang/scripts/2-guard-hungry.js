
const nouns = ["quest", "hunger", "hungry", "belly", "food"]

function onCommand(cmd, rest, mob, room, eventDetails) {
    if (cmd == "wave") {
        mob.Command("wave")
    }
    return false;
}

function onAsk(mob, room, eventDetails) {

    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }

    match = UtilFindMatchIn(eventDetails.askText, nouns);
    if ( match.found ) {

        mob.Command("emote rubs his belly.")
        mob.Command("say I forgot my lunch today, and I'm so hungry.")
        mob.Command("say Do you think you could find a cheese sandwich for me?")

        user.GiveQuest("4-start")

        return true;
    }

    return false;
}

function onGive(mob, room, eventDetails) {

    if (eventDetails.sourceType == "mob") {
        return false;
    }

    if ( eventDetails.gold > 0 ) {
        mob.Command("say I don't need your money... but I'll take it!")
        
        // Check a random number
        if ( Math.random() > 0.5 ) {
            mob.Command("emote flips a coin into the air and catches it!")
        } else {
            mob.Command("emote flips a coin into the air and misses the catch!")
            mob.Command("drop 1 gold");
        }
        return true;
    }

    if (eventDetails.item) {
        if (eventDetails.item.ItemId != 30004) {
            mob.Command("look !"+String(eventDetails.item.ItemId))
            mob.Command("drop !"+String(eventDetails.item.ItemId), UtilGetSecondsToTurns(5))
            return true;
        }
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


// Invoked once every round if mob is idle
function onIdle(mob, room) {

    round = UtilGetRoundNumber();

    grumbled = false
    userIds = room.GetPlayers();

    playersTold = mob.GetTempData('playersTold');
    if ( playersTold === null ) {
        playersTold = {};
    }

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

        if ( Object.keys(playersTold).length > 0 ) {
            mob.SetTempData('playersTold', playersTold);
        } else {
            mob.SetTempData('playersTold', null);
        }
        
        return true;
    }

    sizeBefore = Object.keys(playersTold).length;
    for (var key in playersTold) {
        if ( playersTold[key] < round-100 ) {
            delete playersTold[key];
        }
    }
    sizeAfter = Object.keys(playersTold).length;

    if ( sizeAfter != sizeBefore ) {
        if ( sizeAfter == 0 ) {
            mob.SetTempData('playersTold', playersTold);
        }
    } else {
        mob.SetTempData('playersTold', null);
    }

    action = round % 3;

    if ( action == 0 ) {
        mob.Command("wander")
        return true;
    }

    return false;
}
