
function onCommand_use(user, item, room) {
    
    SendUserMessage(user.UserId(), "You! use the <ansi fg=\"itemname\">Sleeping Bag</ansi>.");
    SendRoomMessage(room.RoomId(), user.GetCharacterName()+" uses their <ansi fg=\"itemname\">Sleeping Bag</ansi>.", user.UserId())

    user.CancelBuffWithFlag("hidden"); // cancel any hidden buff (most item use should do this if it's noticeable)

    user.GiveBuff(15); // Give the sleeping buff
    
    item.AddUsesLeft(-1); // Decrement the uses left by 1
    item.MarkLastUsed(); // Update the last used round number to current

    return true;
}
