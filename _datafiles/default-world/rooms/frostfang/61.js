
function onCommand_sleep(rest, user, room) {

    SendUserMessage(user.UserId(), "You are directed to a room upstairs with a large bed. How inviting...");
    SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" says something to the Inn keeper and is escorted to another room.", user.UserId());
    
    user.MoveRoom(432);
    user.GiveBuff(15);

    return true;
}
