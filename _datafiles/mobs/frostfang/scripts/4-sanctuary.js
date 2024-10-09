
const asksubjects = ["quest", "bishop", "arch-bishop", "king"]

function onAsk(mob, room, eventDetails) {

    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }

    match = UtilFindMatchIn(eventDetails.askText, asksubjects);
    if ( !match.found ) {
        return false;
    }

    mob.Command("say I often see some priests snooping around in the alley behind the Sanctuary.")
    mob.Command("say I used to think they were just taking care of the rat problem, but now I'm not so sure.")

    user.GiveQuest("2-catacombs")

    return true;
}

function onGive(mob, room, eventDetails) {

    if (eventDetails.gold > 0) {

        if ( (user = GetUser(eventDetails.sourceId)) == null ) {
            return false;
        }

        totalDonated = user.GetMiscCharacterData("donation-tally-sanctuary");
        if (totalDonated == null) {
            totalDonated = 0;
        }

        totalDonated += eventDetails.gold;

        user.SetMiscCharacterData("donation-tally-sanctuary", totalDonated);

        mob.Command("say Thank you for your donation.")
        mob.Command("emote nods softly.")

        mob.Command("say You have donated a total of <ansi fg=\"gold\">"+String(totalDonated)+" gold</ansi>!")

        if ( totalDonated >= 200 ) {
            // Give the spell "heal"   
            if ( user.LearnSpell("heal") ) {
                mob.Command("say Thank you for supporting the Sanctuary of the Benevolent Heart!")
                mob.Command("say I'll teach you the <ansi fg=\"spell-helpful\">Heal</ansi> spell. It can cure the gravest of wounds.")
                mob.Command("emote Shows you some useful gestures.")
                mob.Command("say Check your <ansi fg=\"command\">spellbook</ansi>.")
            }

        }
        if ( totalDonated >= 1000 ) {
            // Give the spell "healall"
            if ( user.LearnSpell("healall") ) {
                mob.Command("say Thank you for your extreme support of the Sanctuary of the Benevolent Heart!")
                mob.Command("say I'll teach you the <ansi fg=\"spell-helpful\">Heal All</ansi> spell. It can affect your whole party!")
                mob.Command("emote Shows you some useful gestures.")
                mob.Command("say Check your <ansi fg=\"command\">spellbook</ansi>.")
            }
        }

        return true;
    }
    return false;
}

function onIdle(mob, room) {

    round = UtilGetRoundNumber();
    action = round % 6;

    if (action > 1) {
        return false;
    }

    mob.Command("say Have you taken the time to help the poor today?")

    mob.Command(`say To make a donation, simply <ansi fg="command">give</ansi> the gold to me.`)

    return true;
}