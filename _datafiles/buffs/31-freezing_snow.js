
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
    SendUserMessage(actor.UserId(), "<ansi fg=\"51\">The cold bites for <ansi fg=\"damage\">"+String(harmAmt)+" damage</ansi>!</ansi>\n");
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' is freezing.', actor.UserId());


}