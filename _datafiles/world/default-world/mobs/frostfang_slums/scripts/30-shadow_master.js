

const AMETHYST_ITEM_ID = 5;
const AMETHYST_ROOM_ID = 433;

const ASK_SUBJECTS = ["amethyst", "heist", "bank", "spell", "rats"];

function onAsk(mob, room, eventDetails) {

    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }

    match = UtilFindMatchIn(eventDetails.askText, ASK_SUBJECTS);
    if ( match.found ) {

        mob.Command("say Look, we haven't had issues with rats ever since we discovered how to control them.");
        mob.Command("say If you can recover the amethyst from the bank vault for us, I'll teach you too.");
        
        return true;
    }

    match = UtilFindMatchIn(eventDetails.askText, lichSubjects);
    if ( match.found ) {
        mob.Command("say An ancient lich king eh? Do you have any proof that what you say is true?");

        return true;
    }

    return true;
}


function onGive(mob, room, eventDetails) {

    if (eventDetails.sourceType == "mob") {
        return false;
    }

    if ( eventDetails.gold > 0 ) {
        mob.Command("say I'll use this money to buy more books!")
        return true;
    }

    if (eventDetails.item) {
        if (eventDetails.item.ItemId != 5) {
            mob.Command("look !"+String(eventDetails.item.ItemId))
            mob.Command("drop !"+String(eventDetails.item.ItemId), UtilGetSecondsToTurns(5))
            return true;
        }
    }

    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }


    mob.Command("say Well done! A deal's a deal!")
    mob.Command("say I'll teach you the <ansi fg=\"spell-helpful\">Charm Rat</ansi> spell.")
    mob.Command("emote Shows you some useful gestures.")
    mob.Command("say Check your <ansi fg=\"command\">spellbook</ansi>.")

    user.LearnSpell("charmrat")

    return true;
}
