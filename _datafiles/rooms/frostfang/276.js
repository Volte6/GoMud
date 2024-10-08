
locketAvailableRound = 0;

const glimmer = ["leaves", "glimmer", "light", "locket"];
const locket = ["gold", "golden", "golden locket", "locket", "sophie's locket", "object"];
const verbs = ["get", "take", "grab"];

function onCommand_look(rest, user, room) {

    roundNow = UtilGetRoundNumber();
    if ( roundNow < locketAvailableRound ) {
        return false;
    }

    if ( user.HasQuest("1-return") ) {
        return false;
    }

    matches = UtilFindMatchIn(rest, glimmer);
    if ( matches.found ) {
        SendUserMessage(user.UserId(), "Nestled inside a pile of leaves is some sort of golden object.");
        SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" seems to be digging around in the leaves.", user.UserId());   
        return true;
    }

    matches = UtilFindMatchIn(rest, locket);
    if ( matches.found ) {
        SendUserMessage(user.UserId(), "It appears to be a <ansi fg=\"itemname\">golden locket</ansi>.");
        return true;
    }

    return false;
}

// Generic Command Handler
function onCommand(cmd, rest, user, room) {

    roundNow = UtilGetRoundNumber();
    if ( roundNow < locketAvailableRound ) {
        return false;
    }

    if ( !verbs.includes(cmd) ) {
        return false;
    }

    if ( user.HasQuest("1-return") ) {
        return false;
    }

    matches = UtilFindMatchIn(rest, locket);
    if ( !matches.found ) {
        return false;
    }
    
    SendUserMessage(user.UserId(), "You brush aside the leaves and take the <ansi fg=\"itemname\">golden locket</ansi>.");
    SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" takes a <ansi fg=\"itemname\">golden locket</ansi> from the pile of leaves.", user.UserId());
    
    user.GiveItem(20025);
    user.GiveQuest("1-return");

    locketAvailableRound = roundNow + UtilGetMinutesToRounds(15);

    return true;
}



