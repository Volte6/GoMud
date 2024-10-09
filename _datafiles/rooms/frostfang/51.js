
const lantern = ["lantern", "light"];
const verbs = ["touch", "fix", "pull", "light", "take", "get", "move", "adjust", "turn", "rub", "clean", "dust", "wipe", "polish", "repair", "break", "smash", "ignite", "light"];


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
    SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" does something with the lantern, and a secret passage opens! He disappears through the passage just as it closes.", user.UserId());
    user.MoveRoom(50);

    return true;
}



