
const startNouns = ["rat", "rats", "too many", "problem"];
const thievesNouns = ["thief", "thieves", "guild", "hideout", "entrance", "dogs", "slums"];

function onAsk(mob, room, eventDetails) {

    user = GetUser(eventDetails.sourceId);

        // Waiting for a player to ask about the rats
        if ( !user.HasQuest("7-start") ) {
            startMatch = UtilFindMatchIn(eventDetails.askText, startNouns);
            if ( startMatch.found ) {           
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

        tMatch = UtilFindMatchIn(eventDetails.askText, thievesNouns);
        if ( tMatch.found ) { 

            askTimes = mob.GetTempData(`ask-`+String(user.UserId()));
            if ( askTimes == null ) {
                askTimes = 0;
                
            }
            
            askTimes++;
            mob.SetTempData(`ask-`+String(user.UserId()), askTimes);

            if ( askTimes == 1 ) {
                mob.Command("say I really shouldn't talk about the thieves guild. They like it secret for a reason, and it could mean my neck!");
                return true;
            }

            if ( askTimes > 1 ) {
                mob.Command("say Okay look, their entrance is near the far south end of the slums.");
                mob.Command("say There are some dogs guarding it, and it takes a little searching around to find the hidden entrance.");
                return true;
            }

        } else {

            mob.Command("say Thank you so much! I can finally get back to catching some rats, and maybe earn a little coin.");
            mob.Command("say The thieves guild used to employ me to eliminate rats around their hideout, but for some reason they don't seem to need my help anymore, and didn't pay me for my last job I did for them.");

            return true;
        }

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

        mob.Command("say Thank you so much! I can finally get back to catching some rats, and maybe earn a little coin.");
        mob.Command("say The thieves guild used to employ me to eliminate rats around their hideout, but for some reason they don't seem to need my help anymore, and didn't pay me for my last job I did for them.");
        
        user.GiveQuest("7-end");

    }


}


RANDOM_IDLE = [
    "emote shakes his head in disbelief.",
    "emote attempts to fix a rat trap.",
    "say There's just too many rats. We'll never get rid of them all.",
    "say I'm so tired. I need a break.",
    "say I'm running out of traps. I need to find more.",
    "say I'm worried about the rats in the slums. They're everywhere!",
    "say I'm running out of traps and don't seem to be making a dent in the rat numbers."
];

// Invoked once every round if mob is idle
function onIdle(mob, room) {

    if ( UtilGetRoundNumber()%3 != 0 ) {
        return true;
    }

    randNum = UtilDiceRoll(1, 10)-1;
    if ( randNum < RANDOM_IDLE.length ) {
        mob.Command(RANDOM_IDLE[randNum]);
        return true;
    }

    mob.Command("wander");

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
