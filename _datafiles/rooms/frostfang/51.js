
const lantern = ["lantern", "light"];
const verbs = ["touch", "fix", "pull", "light", "take", "get", "move", "adjust", "turn", "rub", "clean", "dust", "wipe", "polish", "repair", "break", "smash", "ignite", "light"];

function onCommand_look(rest, user, room) {

    matches = UtilFindMatchIn(rest, lantern);
    if ( !matches.found ) {
        return false;
    }

    SendUserMessage(user.UserId(), "The lantern is old, and doesn't appear to be in working order. It's hanging on the wall, but the glass is cracked and the wick is missing.");
    SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" examines an old lantern on the wall.");   
    
    return true;
}

// Generic Command Handler
function onCommand(cmd, rest, user, room) {

    if ( !verbs.includes(cmd) ) {
        return false;
    }

    matches = UtilFindMatchIn(rest, lantern);
    if ( !matches.found ) {
        return false;
    }

    SendUserMessage(user.UserId(), "You adjust the latern, and a passage way reveals itself in the wall. Just as you step through the passage, it closes behind you.");
    SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" does something with the lantern, and a secret passage opens! He disappears through the passage just as it closes.");
    user.MoveRoom(50);

    return true;
}



