

const nouns = ["sarcophagus", "tomb"];
const verbs = ["touch", "push", "hit", "kick", "open", "pry"];


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
    SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" has triggered some sort of trap! the room begins to tremble and a trap door opens beneath your feet. You fall into the darkness below.", user.UserId());


    players = room.GetPlayers();
    for (var i = 0; i < players.length; i++) {
        if ( (user = GetUser(players[i])) !== null ) {
            SendRoomMessage(138, user.GetCharacterName(true)+" falls into the room from above.", user.UserId());
            user.MoveRoom(138);
        }
    }
    
    return true;
}


      