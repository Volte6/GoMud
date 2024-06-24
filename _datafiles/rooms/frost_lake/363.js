
mapSignData = ""

// Generic Command Handler
function onCommand(cmd, rest, user, room) {

    tryBoat = false;

    boatMatches =  UtilFindMatchIn(cmd, ['boat']);
    if ( cmd == `b` || cmd == `bo` || boatMatches.found ) {

        if ( !user.HasItemId(10016) ) {
            SendUserMessage(user.UserId(), "The boats have no oars, and can't be rowed or paddled.");
            SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" scratches their head while examining the boat.");
            
            return true;
        }

        SendUserMessage(user.UserId(), "You pull out your oar and use it to paddle across the water.");
        SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" pulls out an oar paddles across the water.");
    }

    return false;
}
