
function onCommand_use(user, item, room) {
    
    
    SendUserMessage(user.UserId(), "You thrust your fist containing the <ansi fg=\"itemname\">"+item.Name()+"</ansi> into the air. Suddenly it bursts into a shower of golden sparks.");
    SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" thrusts their fist containing a <ansi fg=\"itemname\">"+item.Name()+"</ansi> into the air. Suddenly it bursts into a shower of golden sparks.", user.UserId())

    SendUserMessage(user.UserId(), "");
    SendUserMessage(user.UserId(), "<ansi fg=\"yellow-bold\">*******************************************************************************</ansi>");
    SendUserMessage(user.UserId(), "");
    SendUserMessage(user.UserId(), "<ansi fg=\"yellow-bold\">You just gained a STAT POINT!</ansi>");
    SendUserMessage(user.UserId(), "");
    SendUserMessage(user.UserId(), "<ansi fg=\"yellow-bold\">*******************************************************************************</ansi>");

    item.AddUsesLeft(-1); // Decrement the uses left by 1
    item.MarkLastUsed(); // Update the last used round number to current
    user.GiveStatPoints(1);

    return true;
}
