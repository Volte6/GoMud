
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    
    quarryUserName = actor.GetMiscCharacterData("tracking-user");
    quarryMobName = actor.GetMiscCharacterData("tracking-mob");

    if ( quarryUserName != null ) {
        SendUserMessage(actor.UserId(), 'Your senses are heightened as you focus your tracking skills on <ansi fg="username">'+quarryUserName+'</ansi>.');
    } else {
        SendUserMessage(actor.UserId(), 'Your senses are heightened as you focus your tracking skills on <ansi fg="mobname">'+quarryMobName+'</ansi>.');
    }

}

// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {

    quarryUserName = actor.GetMiscCharacterData("tracking-user");
    quarryMobName = actor.GetMiscCharacterData("tracking-mob");

    if ( quarryUserName != null ) {
        SendUserMessage(actor.UserId(), 'You are no longer actively tracking <ansi fg="username">'+quarryUserName+'</ansi>.');
    } else {
        SendUserMessage(actor.UserId(), 'You are no longer actively tracking <ansi fg="mobname">'+quarryMobName+'</ansi>.');
    }

    actor.SetMiscCharacterData("tracking-mob", null);
    actor.SetMiscCharacterData("tracking-user", null);

    

}
