
const startNouns = ["rat", "rats", "too many", "problem"];

function onAsk(mob, room, eventDetails) {

    user = GetUser(eventDetails.sourceId);

        // Waiting for a player to ask about the rats
        if ( !user.HasQuest("7-start") ) {
            startMatch = UtilFindMatchIn(eventDetails.askText, startNouns);
            if ( startMatch.exact.length > 0 ) {           
                    mob.Command("say I'm worried about the rats in the slums. They're everywhere!");
                    mob.Command("say I'm running out of traps and don't seem to be making a dent in the rat numbers.");
                    mob.Command("say If you can kill 25 of them, come back and see me. I'll pay you for your trouble.");
    
                    user.GiveQuest("7-start");
                    return true;
            }
            return false;
        }

    // Waiting for players to show him 25 rodent kills
    if ( user.HasQuest("7-start") && !user.HasQuest("7-gettrap") ) {
        ratkillCt = user.GetRaceKills("rodent"); 

        if ( ratkillCt >= 25 ) {
            mob.Command("say Thank you for killing those rats! I can finally get a little rest.");
            mob.Command("say While you're feeling helpful, if you could recover a rat trap from a frostfang citizen I was working for, I would be very grateful.");
            
            user.GiveQuest("7-gettrap");
            return true;
        }

        mob.Command("say Looks like you've killed <ansi fg=\"red\">"+String(ratkillCt)+" rats</ansi>. Keep up the good work! Remember you can type <ansi fg=\"command\">kills</ansi> to check your progress.");
        return true;
    }

    if ( user.HasQuest("7-end") ) {
        mob.Command("say I'll let you in on a little secret... the thieves guild can be found in the far southern part of the slums.");
        mob.Command("say There are some dogs guarding the entrance, but if you can search around that area, you'll find the entrance!");
    }

    return false;
}

function onGive(mob, room, eventDetails) {

    if (eventDetails.sourceType == "mob") {
        return false;
    }

    if ( eventDetails.gold > 0 ) {
        mob.Command("say Ah, a tip! Much appreciated!");
        return true;
    }

    if (eventDetails.item) {
        if (eventDetails.item.ItemId != 11) {
            mob.Command("look !"+String(eventDetails.item.ItemId))
            mob.Command("drop !"+String(eventDetails.item.ItemId), UtilGetSecondsToTurns(5))
            return true;
        }


        user = GetUser(eventDetails.sourceId);

        mob.Command("say Thank you so much! I can finally get back to catching some rats.");
        mob.Command("say Here's a little secret... The thieves guild used to employ me to eliminate rats around their hideout.");
        mob.Command("say For some reason they don't seem to need my help anymore, and didn't pay me for my last job I did for them.");
        mob.Command("say So i'll let you in on a little secret... they can be found in the far southern part of the slums.");
        mob.Command("say There are some dogs guarding the entrance, but if you can search around that area, you'll find the entrance!");
        
        user.GiveQuest("7-end");

    }


}


// Invoked once every round if mob is idle
function onIdle(mob, room) {

    round = UtilGetRoundNumber();

    action = round % 4;

    if ( action == 0 ) {
        mob.Command("emote shakes his head in disbelief.");
        mob.Command("say There's just too many rats. We'll never get rid of them all.");
        return true;
    }

    return false;
}


function onShow(mob, room, eventDetails) {

    if (eventDetails.item.ItemId == 11) {
        
        mob.Command("say Perfect! Give it to me, please.");
        return true;

    }

    mob.Command("emote is uninterested.");    
    return false;
}