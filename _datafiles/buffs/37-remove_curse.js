
// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {

    items = actor.Uncurse();
    room = GetRoom(actor.GetRoomId())

    for( var i in items ) {

        actor.SendText(`<ansi fg="buff-text">You feel a curse lifted from your `+items[i].Name()+`.</ansi>`)
        
        message = `<ansi fg="buff-text">The `+items[i].Name()+` held by `+actor.GetCharacterName(true)+` glows briefly.</ansi>`;
        
        room.SendText(message, actor.UserId())

    }

}
