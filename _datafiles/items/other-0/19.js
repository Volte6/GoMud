
function onCommand_read(user, item, room) {

    SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" thumbs through their <ansi fg=\"item\">"+item.DisplayName()+"</ansi> book.", user.UserId());   

    if ( user.LearnSpell("curepoison") ) {
        SendUserMessage(user.UserId(), "You discover the the <ansi fg=\"spell-helpful\">Cure Poison</ansi> spell. It can remove a deadly ailment.");
        SendUserMessage(user.UserId(), "Check your <ansi fg=\"command\">spellbook</ansi>.");
    }

    return true;
}

