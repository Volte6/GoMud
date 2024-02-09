
const lantern = ["lantern", "light"];
const verbs = ["touch", "fix", "pull", "light", "take", "get", "move", "adjust", "turn", "rub", "clean", "dust", "wipe", "polish", "repair", "break", "smash", "ignite", "light"];

function onCommand_look(rest, userId, roomId) {


    parts = rest.toLowerCase().split(' ');
    for (var i = 0; i < parts.length; i++) {
        matches = UtilFindMatchIn(parts[i], lantern);
        if ( matches.exact.length > 0  ) {
            SendUserMessage(userId, "The lantern is old, and doesn't appear to be in working order. It's hanging on the wall, but the glass is cracked and the wick is missing.");
            SendRoomMessage(roomId, "<ansi fg=\"username\">"+UserGetCharacterName(userId)+"</ansi> examines an old lantern on the wall.");   
            return true;
        }
    }

    return false;
}

// Generic Command Handler
function onCommand(cmd, rest, userId, roomId) {

    if ( !verbs.includes(cmd) ) {
        return false;
    }

    roundNow = UtilGetRoundNumber();

    parts = rest.toLowerCase().split(' ');
    for (var i = 0; i < parts.length; i++) {
        matches = UtilFindMatchIn(parts[i], lantern);
        if ( matches.exact.length > 0  ) {
            
            SendUserMessage(userId, "You adjust the latern, and a passage way reveals itself in the wall. Just as you step through the passage, it closes behind you.");
            SendRoomMessage(roomId, "<ansi fg=\"username\">"+UserGetCharacterName(userId)+"</ansi> does something with the lantern, and a secret passage opens! He disappears through the passage just as it closes.");
            UserMoveRoom(userId, 50);

            return true;
        }
    }

    return false;
}



