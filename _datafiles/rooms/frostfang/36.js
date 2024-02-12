

function onCommand_north(rest, user, room) {


    hasQuestUserIds = room.HasQuest("2-start", user.UserId());
    if ( hasQuestUserIds.length < 1 ) {

        SendUserMessage(user.UserId(), 'The guards block your path. "<ansi fg="yellow">You must be invited to enter the throne room</ansi>," they say. "<ansi fg="yellow">We cannot let you pass.</ansi>"');
        SendRoomMessage(room.RoomId(), '<ansi fg="username">'+user.GetCharacterName()+'</ansi> tries to enter the throne room, but the guards block the way.');
        
        return true;
    }

    return false;

}