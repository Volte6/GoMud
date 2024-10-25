
function onCommand_sweep(user, item, room) {

    SendUserMessage(user.UserId(), "You sweep the floors thoroughly, until not a single dust bunny can be found.");
    SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" sweeps their heart out with <ansi fg=\"item\">"+item.Name(true)+"</ansi>.", user.UserId());   

    room.RemoveMutator("dusty-floors")

    return true;
}

