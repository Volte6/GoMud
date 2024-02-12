
const asksubjects = ["quest", "bishop", "arch-bishop", "king"]

function onAsk(mob, room, eventDetails) {

    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }

    match = UtilFindMatchIn(eventDetails.askText, asksubjects);
    if ( match.exact.length < 1 ) {
        return false;
    }

    mob.Command("say I often see some priests snooping around in the alley behind the Sanctuary.")
    mob.Command("say I used to think they were just taking care of the rat problem, but now I'm not so sure.")

    user.GiveQuest("2-catacombs")

    return true;
}

function onGive(mob, room, eventDetails) {

    if (eventDetails.gold > 0) {

        if (eventDetails.gold < 100) {
            mob.Command("say Thank you for your donation.")
            mob.Command("emote nods softly.")
        } else {
            mob.Command("say Thank you for your generous donation.")
            mob.Command("emote bows deeply.")
            mob.Command("emote claps his hands together in prayer and begins to chant.")
            mob.Command("uncurse")
        }
        return true;
    }
    return false;
}

function onIdle(mob, room) {

    round = UtilGetRoundNumber();
    action = round % 6;

    if (action > 2) {
        return false;
    }

    if (action == 2 ) {
        mob.Command("say Have you taken the time to help the poor today?")
    } else if (action == 1) {
        mob.Command(`say For a <ansi fg="gold">100 gold</ansi> donation, I will remove any curses afflicting your party.`)
    }

    mob.Command(`say To make a donation, simply <ansi fg="command">give</ansi> the gold to me.`)

    return true;
}