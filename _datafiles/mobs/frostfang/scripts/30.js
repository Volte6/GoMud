

const AMETHYST_ITEM_ID = 5;
const AMETHYST_ROOM_ID = 433;
const TRAINEE_MOB_ID = 41;

const ASK_SUBJECTS = ["amethyst", "heist", "bank", "spell", "rats"];

const RANDOM_IDLE = [
    "say You have done well to make it this far into my domain. Perhaps you should train here.",
    "say That pesky rat catcher keeps trying to get more work from us. I told him we don't need him anymore.",
];

const RANDOM_CONVERSE = [
    "converse have you had any luck locating the amethyst?"
]

function onAsk(mob, room, eventDetails) {

    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }

    match = UtilFindMatchIn(eventDetails.askText, ASK_SUBJECTS);
    if ( match.found && match.exact.length > 0 ) {

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

// Invoked once every round if mob is idle
function onIdle(mob, room) {

    if ( UtilGetRoundNumber()%3 > 0 ) {
        return true;
    }

    randomNum = UtilDiceRoll(1, 2);

    traineePresent = false;
    mobList = room.GetMobs();
    for ( var k in mobList ) {
        tmpMob = GetMob(mobList[k])
        if ( tmpMob.MobTypeId() == TRAINEE_MOB_ID ) {
            traineePresent = true;
            break;
        }
    }


    // 1/3 chance of defaulting to regular idle
    if ( !traineePresent || randomNum == 1 ) {
        randNum = UtilDiceRoll(1, 10)-1;
        if ( randNum < RANDOM_IDLE.length ) {
            mob.Command(RANDOM_IDLE[randNum]);
        }
        return true;
    }


    randNum = UtilDiceRoll(1, 5)-1;
    if ( randNum < RANDOM_CONVERSE.length ) {
        mob.Command(RANDOM_CONVERSE[randNum]);
        return true;
    }

    return true;
}



CONVERSE_RESPONSES = {
    "yes, it is in the bank vault.": "converse well then, what is the plan to retrieve it?",
    "we learned recently of a secret tunnel below it.": "converse interesting, then lets talk more about it later in private."
}

function onConverse(message, mob, sourceMob, room) {

    if ( sourceMob.MobTypeId() != TRAINEE_MOB_ID ) {
        return false;
    }

    if ( CONVERSE_RESPONSES[message] != null ) {
        mob.Command(CONVERSE_RESPONSES[message], UtilGetSecondsToTurns(2));
        return true;
    }
}
