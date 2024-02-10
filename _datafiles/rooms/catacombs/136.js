

const nouns = ["sarcophagus", "tomb"];
const verbs = ["touch", "push", "hit", "kick", "open", "pry"];

function onCommand_look(rest, user, room) {

    matches = UtilFindMatchIn(rest, nouns);
    if ( !matches.found ) {
        return false;
    }

    SendUserMessage(user.UserId(), "The sarcophagus lies solemn and imposing, its ancient stone surface etched with enigmatic runes and the stern visage of the entombed sovereign, exuding an air of timeless dominion and whispered dread.");
    SendRoomMessage(room.RoomId(), "<ansi fg=\"username\">"+user.GetCharacterName()+"</ansi> examines the sarcophagus.");   

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

    SendUserMessage(user.UserId(), "The room begins to tremble, and a trap door opens beneath your feet! You fall into the room below!");
    SendRoomMessage(room.RoomId(), "<ansi fg=\"username\">"+user.GetCharacterName()+"</ansi> has triggered some sort of trap! the room begins to tremble and a trap door opens beneath your feet. You fall into the darkness below.");


    players = room.GetPlayers();
    for (var i = 0; i < players.length; i++) {
        if ( (user = GetUser(players[i])) !== null ) {
            SendRoomMessage(138, "<ansi fg=\"username\">"+user.GetCharacterName()+"</ansi> falls into the room from above.");
            user.MoveRoom(138);
        }
    }
    
    return true;
}


      