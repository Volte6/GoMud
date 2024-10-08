

function onCommand_north(rest, user, room) {


    hasQuestUserIds = room.HasQuest("2-start", user.UserId());
    if ( hasQuestUserIds.length < 1 ) {

        SendUserMessage(user.UserId(), 'The guards block your path. "<ansi fg="yellow">You must be invited to enter the throne room</ansi>," they say. "<ansi fg="yellow">We cannot let you pass.</ansi>"');
        SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+' tries to enter the throne room, but the guards block the way.', user.UserId());
        
        return true;
    }

    return false;

}