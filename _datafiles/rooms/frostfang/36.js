
function onCommand_north(rest, user, room) {

    if ( !user.HasQuest("2-start") ) {
        SendUserMessage(user.UserId(), "The guards block your path. \"You must be invited to enter the throne room,\" they say. \"We cannot let you pass.\"");
        return true;
    }
    
    return false;
}
