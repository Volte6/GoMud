
const altar = ["altar"];

function onCommand_look(rest, user, room) {

    matches = UtilFindMatchIn(rest, altar);
    if ( matches.found ) {
        SendUserMessage(user.UserId(), "<ansi fg=\"240\">The smell of the insense fills your nostrels, numbing your senses.</ansi>");       
        user.GiveBuff(2, "drugs");
        return true;
    }

    return false;
}
