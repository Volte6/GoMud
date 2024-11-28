
// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {

    items = actor.Uncurse();
    room = GetRoom(actor.GetRoomId())

    for( var i in items ) {

        actor.SendText(`You feel a curse lifted from your `+items[i].Name())
        
        message = `The `+items[i].Name()+` held by `+actor.GetCharacterName(true)+` glows briefly.`;
        
        room.SendText(message, actor.UserId())

    }

}
