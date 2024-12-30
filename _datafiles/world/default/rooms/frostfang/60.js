
function onCommand_west(rest, user, room) {

    if ( !UtilIsDay() ) {
        SendUserMessage(user.UserId(), "The eastern city gates close every night. You'll have to wait for day, or find another way in.");
        return true;
    }
    
    return false;
}
