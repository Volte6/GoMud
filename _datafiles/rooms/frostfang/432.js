
crowbarAvailableRound = 0;

const crowbar = ["crowbar", "rod", "metal", "bar"];
const verbs = ["get", "take", "grab", "steal", "snatch"];

function onCommand_look(rest, user, room) {

    matches = UtilFindMatchIn(rest, crowbar);
    if ( !matches.found ) {
        return false;
    }

    roundNow = UtilGetRoundNumber();

    if (roundNow < crowbarAvailableRound) {
        return false;
    }

    SendUserMessage(user.UserId(), "A <ansi fg=\"item\">crowbar</ansi> leans besides the fireplace. Probably used for poking the fire and moving the logs around.");
    SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" looks at the <ansi fg=\"item\">crowbar</ansi> by the fireplace.");   

    return true;
}

// Generic Command Handler
function onCommand(cmd, rest, user, room) {

    if ( !verbs.includes(cmd) ) {
        return false;
    }
    
    matches = UtilFindMatchIn(rest, crowbar);
    if ( !matches.found ) {
        return false;
    }

    roundNow = UtilGetRoundNumber();
    
    if (roundNow < crowbarAvailableRound) {
        return false;
    }

    crowbarAvailableRound = roundNow + UtilGetMinutesToRounds(15)

    SendUserMessage(user.UserId(), "You take the <ansi fg=\"item\">crowbar</ansi>. They probably won't miss it.");
    SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" takes the <ansi fg=\"item\">crowbar</ansi> from beside the fireplace.");   
    
    user.GiveItem(10012);

    return true;
}


