
// Generic Command Handler
function onCommand(cmd, rest, user, room) {

    if ( rest.substr(rest.length - 2) != "ob" ) {
        return false;
    }

    if ( cmd == "look" ) {

        SendUserMessage(user.UserId(), "The obelisk crackles with subtle dark energy. It seems like it could be dangerous to the touch.");
        SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" looks at the obelisk.");
        
        return true;
    }

    if ( cmd == "touch" ) {
      
        SendUserMessage(user.UserId(), "You reach out and touch the obelisk.");
        SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" reaches out and touches the obelisk.");

        if ( !user.TrainSkill("portal", 1) ) {
            
            SendUserMessage(user.UserId(), "Nothing happens.");

        }
        
        return true;
    }
    
    return false;
}