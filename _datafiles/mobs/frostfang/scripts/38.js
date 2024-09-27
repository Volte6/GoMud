
// Invoked once every round if mob is idle
function onIdle(mob, room) {

    roundNow = UtilGetRoundNumber();

    charmedUserId = mob.GetCharmedUserId();

    if ( charmedUserId == 0 ) {
        return false;
    }

    charmer = GetUser(charmedUserId);

    // If charmer isn't in world, skip
    if ( charmer == null ) {
        return false;
    }

    // If they aren't in the same room for some reason, skip
    if ( charmer.GetRoomId() != mob.GetRoomId() ) {

        lostRoundCt = mob.GetTempData(`roundsLost`);
        if (lostRoundCt == null ) lostRoundCt = 0;

        lostRoundCt++;

        if ( lostRoundCt >= 3 ) {
            
            mob.SetTempData('roundsLost', 0);

            targetRoom = GetRoom(charmer.GetRoomId() );
            if ( targetRoom != null ) {
                mob.MoveRoom( charmer.GetRoomId() );
                room.SendText(`A large ` + UtilApplyColorPattern('swirling portal', 'pink') + ` appears, and ` + mob.GetCharacterName(true) + ` steps into it, right before disappears.`);
                targetRoom.SendText(`A large ` + UtilApplyColorPattern('swirling portal', 'pink') + ` appears, and ` + mob.GetCharacterName(true) + ` steps out of it, right before disappears.`);
                mob.Command(`say I almost lost you ` + charmer.GetCharacterName(true) + `!`);
            }
            
            return true;
        }
        
        mob.SetTempData('roundsLost', lostRoundCt);

        return false;

    }
    
    lastTipRound = mob.GetTempData(`lastTipRound`);
    if ( lastTipRound == null ) {
        lastTipRound = 0;
    }

    lastUserInput = charmer.GetLastInputRound();
    roundsSinceInput = roundNow - lastUserInput;

    // Only give a tip if the user has been inactive for 5 rounds
    if ( roundsSinceInput < 3 ) {
        return true;
    }

    roundsPassed = roundNow - lastTipRound;

    // give at least 5 rounds between tips, even if the user remains inactive.
    if ( roundsPassed < 5 ) {
        return false;
    }

    switch( UtilDiceRoll(1, 10) ) {
        case 1:
            if ( charmer.GetStatPoints() > 0 ) {
                mob.Command(`sayto @` + charmer.UserId() + ` It looks like you've got some stat points to spend. Type <ansi fg="command">status train</ansi> to upgrade your stats!`);
                mob.SetTempData(`lastTipRound`, roundNow);
            }
            break;
        case 2:
            if ( !charmer.HasQuest(`4-start`) ) {
                mob.Command(`sayto @` + charmer.UserId() + ` There's a guard in the barracks that constantly complains about being hungry. You should <ansi fg="command">ask</ansi> him about it.`);
                mob.SetTempData(`lastTipRound`, roundNow);
            }
            break;
        case 3:
            if ( !charmer.HasQuest(`2-start`) ) {
                mob.Command(`sayto @` + charmer.UserId() + ` I have heard the king worries. If we can find an audience with him we can try to <ansi fg="command">ask</ansi> him about a quest. He is north of town square.`);
                mob.SetTempData(`lastTipRound`, roundNow);
            }
            break;
        case 4:
            mob.Command(`sayto @` + charmer.UserId() + ` There are some rats to bash around the temple south of town square. Just don't go TOO far south, it get dangerous!`);
            mob.SetTempData(`lastTipRound`, roundNow);
            break;
        case 5:
            mob.Command(`sayto @` + charmer.UserId() + ` You can find help on many subjects by typing <ansi fg="command">help</ansi>.`);
            mob.SetTempData(`lastTipRound`, roundNow);
            break;
        default:
            mob.SetTempData(`lastTipRound`, roundNow);
            break;
    }

    return true;
}
