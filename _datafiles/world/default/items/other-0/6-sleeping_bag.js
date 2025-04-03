
function onCommand_use(user, item, room) {
    
    SendUserMessage(user.UserId(), "You unroll the <ansi fg=\"itemname\">"+item.Name()+"</ansi> and hop in.");
    SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" unrolls their <ansi fg=\"itemname\">"+item.Name()+"</ansi> and crawls inside.", user.UserId())

    user.CancelBuffWithFlag("hidden"); // cancel any hidden buff (most item use should do this if it's noticeable)

    user.GiveBuff(15, "sleep"); // Give the sleeping buff
    
    item.AddUsesLeft(-1); // Decrement the uses left by 1
    item.MarkLastUsed(); // Update the last used round number to current

    return true;
}
