
const verbs = ["touch", "push", "press", "take", "rub", "polish"];
const nouns = ["raven", "eyes", "bird"];


// cmd specific handler
function onCommand_look(rest, user, room) {

    matches = UtilFindMatchIn(rest, nouns);
    if ( !matches.found ) {
        return false;
    }

    SendUserMessage(user.UserId(), "Looking more closely, the eyes of the raven are made of onyx. They are clean and clear, as if polished.")
    SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" examines the raven in the mural.");

    return true;
}

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
    SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" presses in the eyes of the raven, and falls through into a room to the west!");

    user.GiveQuest("2-investigate")
    user.MoveRoom(31)
        
    return true;
}

