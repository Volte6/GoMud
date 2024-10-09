
const verbs = ["touch", "push", "press", "take", "rub", "polish"];
const nouns = ["raven", "eyes", "bird"];

// Generic Command Handler
function onCommand(cmd, rest, user, room) {

    if ( !verbs.includes(cmd) ) {
        return false;
    }
    
    matches = UtilFindMatchIn(rest, nouns);
    if ( !matches.found ) {
        return false;
    }

    SendUserMessage(user.UserId(), "You press the eyes of the raven, and follow a secret entrance to the west!");
    SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" presses in the eyes of the raven, and falls through into a room to the west!", user.UserId());

    user.GiveQuest("2-investigate")
    user.MoveRoom(31)
        
    return true;
}

