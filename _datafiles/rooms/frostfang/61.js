
function onCommand_sleep(rest, userId, roomId) {
    SendUserMessage(userId, "You are directed to a room upstairs with a large bed. How inviting...");
    SendRoomMessage(roomId, "<ansi fg=\"username\">"+UserGetCharacterName(userId)+"</ansi> says something to the Inn keeper and is escorted to another room.");
    
    UserMoveRoom(userId, 432);
    UserGiveBuff(userId, 15);

    return true;
}
