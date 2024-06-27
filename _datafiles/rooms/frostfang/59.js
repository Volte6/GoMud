
function onCommand_east(rest, user, room) {

    if ( !UtilIsDay() ) {
        SendUserMessage(user.UserId(), "The east gates are closed for the night. You'll have to wait for day, or find another way out.");
        return true;
    }
    
    return false;
}
