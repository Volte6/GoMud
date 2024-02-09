
crowbarAvailableRound = 0;

const crowbar = ["crowbar", "rod", "metal", "bar"];
const verbs = ["get", "take", "grab", "steal", "snatch"];

function onCommand_look(rest, userId, roomId) {

    parts = rest.toLowerCase().split(' ');
    for (var i = 0; i < parts.length; i++) {
        matches = UtilFindMatchIn(parts[i], crowbar);
        if ( matches.exact.length > 0  ) {
            break;
        }
    }

    if ( matches.exact.length < 1 ) {
        return false;
    }

    roundNow = UtilGetRoundNumber();

    if (roundNow < crowbarAvailableRound) {
        return false;
    }

    SendUserMessage(userId, "A <ansi fg=\"item\">crowbar</ansi> leans besides the fireplace. Probably used for poking the fire and moving the logs around.");
    SendRoomMessage(roomId, "<ansi fg=\"username\">"+UserGetCharacterName(userId)+"</ansi> looks at the <ansi fg=\"item\">crowbar</ansi> by the fireplace.");   

    return true;
}

// Generic Command Handler
function onCommand(cmd, rest, userId, roomId) {

    if ( !verbs.includes(cmd) ) {
        return false;
    }
    
    parts = rest.toLowerCase().split(' ');
    for (var i = 0; i < parts.length; i++) {
        matches = UtilFindMatchIn(parts[i], crowbar);
        if ( matches.exact.length > 0  ) {
            break;
        }
    }

    if ( matches.exact.length < 1 ) {
        return false;
    }

    roundNow = UtilGetRoundNumber();
    
    if (roundNow < crowbarAvailableRound) {
        return false;
    }

    crowbarAvailableRound = roundNow + UtilGetMinutesToRounds(15)

    SendUserMessage(userId, "You take the <ansi fg=\"item\">crowbar</ansi>. They probably won't miss it.");
    SendRoomMessage(roomId, "<ansi fg=\"username\">"+UserGetCharacterName(userId)+"</ansi> takes the <ansi fg=\"item\">crowbar</ansi> from beside the fireplace.");   
    UserGiveItem(userId, 10012);

    return true;
}


