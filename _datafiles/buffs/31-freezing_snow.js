
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {

    if ( actor.HasBuffFlag("warmed")  ) {
        actor.RemoveBuff(31)
        return
    }
    harmAmt = actor.AddHealth(-1 * UtilDiceRoll(1, 2));
    if (harmAmt < 1 ) {
        harmAmt *= -1;
    }
    SendUserMessage(actor.UserId(),     '<ansi fg="buff-text"><ansi fg="51">The cold bites for <ansi fg="damage">'+String(harmAmt)+' damage</ansi>!</ansi></ansi>\n');
    SendRoomMessage(actor.GetRoomId(),  '<ansi fg="buff-text">'+actor.GetCharacterName(true)+' is freezing.</ansi>', actor.UserId());


}