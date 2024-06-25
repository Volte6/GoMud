

const island = ["rocky island", "island", "rocks"];

function onCommand_look(rest, user, room) {

    matches = UtilFindMatchIn(rest, island);
    if ( matches.found ) {
        SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" peers out into the water.");
        SendUserMessage(user.UserId(), "A large island sits in the middle of the lake, but a closer rocky island sits to the northwest of here.");
        return true;
    }


    return false;
}


function onIdle(room) {

    if ( UtilGetRoundNumber()%30 == 0 ) {
        SendRoomMessage(room.RoomId(), "A huge wave crashes against the shore, but as it receeds, you notice a small path of shallow water you can follow to a large rock island.");
        room.AddTemporaryExit("shallow water", "shallow water", 828, 10);
    }

    return false;
}