// 
// buff zero (0) is a special buff that when naturally expires, 
// will remove the player from the game without zombie status.
//

// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),    '<ansi fg="buff-text">You sit down and begin your meditation.</ansi>' )
    SendUserMessage(actor.UserId(),    '<ansi fg="buff-text">Your meditation must complete without interruption to quit gracefully.</ansi>')
    SendRoomMessage(actor.GetRoomId(), '<ansi fg="buff-text">'+actor.GetCharacterName(true)+' sits down a begins to meditate.</ansi>', actor.UserId())
}

// Invoked every time the buff is triggered (see roundinterval)
function onTrigger(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     '<ansi fg="buff-text">You continue your meditation. <ansi bg="blue"> *' + triggersLeft + ' rounds left* </ansi>.</ansi>' )
    SendRoomMessage(actor.GetRoomId(),   '<ansi fg="buff-text">'+actor.GetCharacterName(true)+' continues meditating.</ansi>', actor.UserId() )
}
