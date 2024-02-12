
const nouns = ["quest", "freezing", "cold", "help"]
const crystal_nouns = ["crystals", "winterfire", "where"]

function onAsk(mob, room, eventDetails) {

    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }

    match = UtilFindMatchIn(eventDetails.askText, nouns);
    if ( match.found ) {

        mob.Command("say I've been waiting for a shipment of winterfire crystals. They should have been here months ago.")
        mob.Command("say I'll never abandon my post! Can you find out what happened to my crystals?")
        
        user.GiveQuest("5-start")

        return true;
    }

    match = UtilFindMatchIn(eventDetails.askText, crystal_nouns);
    if ( match.found ) {

        mob.Command("say The shipment was supposed to come from the far east city of Mystarion. I'm not sure what happened to it.")
        user.GiveQuest("5-lookeast")

        return true;
    }

    return false;
}

function onGive(mob, room, eventDetails) {

    if (eventDetails.sourceType == "mob") {
        return false;
    }

    if ( eventDetails.gold > 0 ) {
        mob.Command("say Moneys no good here, but every now and then I can pay for a little help.")
        return true;
    }

    if (eventDetails.item) {
        if (eventDetails.item.ItemId != 4) {
            mob.Command("say Finally! My winterfire crystal! Thank you so much!")
            user.GiveQuest("4-end")
            return true;
        }
    }

    return false;
}

// Invoked once every round if mob is idle
function onIdle(mob, room) {

    round = UtilGetRoundNumber();

    if ( round % 5 == 0) {
        missingQuest = room.MissingQuest("5-end");
        if ( missingQuest.length > 0 ) {
            mob.Command("emote shivers silently.");
        }
    } else if ( round % 5 == 3 ) {
        missingQuest = room.MissingQuest("5-end");
        if ( missingQuest.length > 0 ) {
            mob.Command("say I'm so c.c.c...cold...");
        }
    }
     

    return false;
}
