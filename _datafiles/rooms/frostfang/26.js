
const verbs = ["touch", "push", "press", "take", "rub", "polish"];
const nouns = ["raven", "eyes", "bird"];


// cmd specific handler
function onCommand_look(rest, userId, roomId) {

    parts = rest.toLowerCase().split(' ');
    for (var i = 0; i < parts.length; i++) {
        matches = UtilFindMatchIn(parts[i], nouns);
        if ( matches.exact.length > 0 ) {
            break;
        }
    }

    if ( matches.exact.length < 1 ) {
        return false;
    }

    SendUserMessage(userId, "Looking more closely, the eyes of the raven are made of onyx. They are clean and clear, as if polished.")
    SendRoomMessage(roomId, "<ansi fg=\"username\">"+UserGetCharacterName(userId)+"</ansi> examines the raven in the mural.");

    return true;
}

// Generic Command Handler
function onCommand(cmd, rest, userId, roomId) {


    if ( !verbs.includes(cmd) ) {
        return false;
    }
    
    parts = rest.toLowerCase().split(' ');
    for (var i = 0; i < parts.length; i++) {
        matches = UtilFindMatchIn(parts[i], nouns);
        if ( matches.exact.length > 0  ) {
            break;
        }
    }

    if ( matches.exact.length < 1 ) {
        return false;
    }
    
    SendUserMessage(userId, "You press the eyes of the raven, and follow a secret entrance to the west!");
    SendRoomMessage(roomId, "<ansi fg=\"username\">"+UserGetCharacterName(userId)+"</ansi> presses in the eyes of the raven, and falls through into a room to the west!");

    UserGiveQuest(userId, "2-investigate")
    UserMoveRoom(userId, 31)
        
    return true;
}

