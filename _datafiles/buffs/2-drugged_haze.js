
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     '<ansi fg="buff-text">You catch a whiff of a strange odor carried by the smoke.</ansi>' )
    SendRoomMessage(actor.GetRoomId(),  '<ansi fg="buff-text">You notice the eyes of '+actor.GetCharacterName(true)+' glaze oover.</ansi>', actor.UserId())
}

// Invoked every time the buff is triggered (see roundinterval)
function onTrigger(actor, triggersLeft) {

    SendUserMessage(actor.UserId(),     '<ansi fg="buff-text">You find yourself in a blissful haze, wanting nothing more than to sit and watch the world go by.</ansi>')
    SendRoomMessage(actor.GetRoomId(),  '<ansi fg="buff-text">'+actor.GetCharacterName(true)+' looks blissfully content.</ansi>', actor.UserId())
}

// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     '<ansi fg="buff-text">You snap out of your haze.</ansi>')
    SendRoomMessage(actor.GetRoomId(),  '<ansi fg="buff-text">'+actor.GetCharacterName(true)+' shakes off the haze and the glaze in their eyes fades away.</ansi>', actor.UserId())
}
