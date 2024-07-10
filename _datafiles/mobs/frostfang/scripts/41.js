
MASTER_MOB_ID = 30;

CONVERSE_RESPONSES = {
    "have you had any luck locating the amethyst?": "converse yes, it is in the bank vault.",
    "well then, what is the plan to retrieve it?": "converse we learned recently of a secret tunnel below it."
}

function onConverse(message, mob, sourceMob, room) {

    
    if ( sourceMob.MobTypeId() != MASTER_MOB_ID ) {
        return false;
    }

    if ( CONVERSE_RESPONSES[message] != null ) {
        mob.Command(CONVERSE_RESPONSES[message],  UtilGetSecondsToTurns(2));
        return true;
    }
}
