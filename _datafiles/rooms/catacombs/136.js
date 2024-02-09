

const nouns = ["sarcophagus", "tomb"];
const verbs = ["touch", "push", "hit", "kick", "open", "pry"];

function onCommand_look(rest, userId, roomId) {

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

    SendUserMessage(userId, "The sarcophagus lies solemn and imposing, its ancient stone surface etched with enigmatic runes and the stern visage of the entombed sovereign, exuding an air of timeless dominion and whispered dread.");
    SendRoomMessage(roomId, "<ansi fg=\"username\">"+UserGetCharacterName(userId)+"</ansi> examines the sarcophagus.");   

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

    

    SendUserMessage(userId, "The room begins to tremble, and a trap door opens beneath your feet! You fall into the room below!");
    SendRoomMessage(roomId, "<ansi fg=\"username\">"+UserGetCharacterName(userId)+"</ansi> has triggered some sort of trap! the room begins to tremble and a trap door opens beneath your feet. You fall into the darkness below.");
    
    players = RoomGetPlayers(roomId);
    for (var i = 0; i < players.length; i++) {
        SendRoomMessage(138, "<ansi fg=\"username\">"+UserGetCharacterName(players[i])+"</ansi> falls into the room from above.");
        UserMoveRoom(players[i], 138);
    }
    


    return true;
}


      