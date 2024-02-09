
function onCommand_north(rest, userId, roomId) {

    if ( !UserHasQuest(userId, "2-start") ) {
        SendUserMessage(userId, "The guards block your path. \"You must be invited to enter the throne room,\" they say. \"We cannot let you pass.\"");
        return true;
    }
    
    return false;
}
