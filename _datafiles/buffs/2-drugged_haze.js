
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'You catch a whiff of a strange odor carried by the smoke.')
    SendRoomMessage(actor.GetRoomId(), 'You notice the eyes of '+actor.GetCharacterName(true)+' glaze oover.', actor.UserId())
}

// Invoked every time the buff is triggered (see roundinterval)
function onTrigger(actor, triggersLeft) {

    SendUserMessage(actor.UserId(), 'You find yourself in a blissful haze, wanting nothing more than to sit and watch the world go by.')
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' looks blissfully content.', actor.UserId())
}

// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), "You snap out of your haze.")
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' shakes off the haze and the glaze in their eyes fades away.', actor.UserId())
}
