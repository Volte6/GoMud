
// Invoked when the buff is first applied to the player.
function onStart(actor) {
    SendUserMessage(actor.UserId(), "You lay your head down and immediately doze off.")
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' is getting some rest.', actor.UserId())
}

// Invoked every time the buff is triggered (see roundinterval)
function onTrigger(actor) {
    healAmt = actor.AddHealth(UtilDiceRoll(3, 8))

    SendUserMessage(actor.UserId(), 'ZZzzz...')
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' snores loudly.', actor.UserId())
}

// Invoked when the buff has run its course.
function onEnd(actor) {
    SendUserMessage(actor.UserId(), "You wake up!")
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' wakes up.', actor.UserId())

    actor.GiveBuff(16) // Well Rested
}
