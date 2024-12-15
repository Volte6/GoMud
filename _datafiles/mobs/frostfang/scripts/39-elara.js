
const nouns = ["quest", "book", "books", "library", "spell", "spells", "worried", "worry"]


function onAsk(mob, room, eventDetails) {

    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }

    if ( user.HasQuest("6-end") ) {
        mob.Command("say Oh, it's nothing.");
        return true;
    }

    match = UtilFindMatchIn(eventDetails.askText, nouns);
    if ( match.found ) {

        mob.Command("emote thinks for a moment.")
        mob.Command("say I took a book called The History of Frostfang out with me weeks ago and can't remember where I left it.")
        mob.Command("say If you can find it for me, I'll teach you a useful spell.")

        user.GiveQuest("6-start")

        return true;
    }

    return false;
}

function onGive(mob, room, eventDetails) {

    if (eventDetails.sourceType == "mob") {
        return false;
    }

    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }
    
    if ( eventDetails.gold > 0 ) {
        mob.Command("say I'll use this money to buy more books!")
        return true;
    }

    if (eventDetails.item) {
        if (eventDetails.item.ItemId != 10) {
            mob.Command("look !"+String(eventDetails.item.ItemId))
            mob.Command("drop !"+String(eventDetails.item.ItemId), UtilGetSecondsToTurns(5))
            return true;
        }
    }

    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }

    if ( user.HasQuest("6-end") ) {
        mob.Command("say Please don't borrow books from the library without permission!")
        return true;
    }

    if ( user.HasQuest("6-start") ) {

        user.GiveQuest("6-end")

        mob.Command("say Thank you! It is such an interesting history. For example, Frostfang used to be called DragonsFang!")
        mob.Command("say I'll teach you the <ansi fg=\"spell-helpful\">Illuminate</ansi> spell. It's useful in dark places.")
        mob.Command("emote Shows you some useful gestures.")
        mob.Command("say Check your <ansi fg=\"command\">spellbook</ansi>.")

        partyMembers = user.GetPartyMembers();
        for( i = 0; i < partyMembers.length; i++ ) {    
            a = partyMembers[i];
            a.LearnSpell("illum")
        }
        

        return true;
    }

}


// Invoked once every round if mob is idle
function onIdle(mob, room) {

    noQuest = room.MissingQuest("6-start");
    if ( noQuest.length < 1 ) {
        return false;
    }

    round = UtilGetRoundNumber();

    action = round % 6;

    if ( action == 0 ) {
        mob.Command("say now where did I leave that book?")
        return true;
    }

    return false;
}
