
function onCommand_out(rest, userId, roomId) {
    
    mobs = RoomGetMobs(roomId);
    if ( mobs.length > 0 ) {
        SendUserMessage(userId, "The way out is block by denizens of the cave.");
        return true;
    }

    return false;
}
