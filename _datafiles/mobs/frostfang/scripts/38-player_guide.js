
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
                room.SendText(`A large ` + UtilApplyColorPattern('swirling portal', 'pink') + ` appears, and ` + mob.GetCharacterName(true) + ` steps into it, right before it disappears.`);
                targetRoom.SendText(`A large ` + UtilApplyColorPattern('swirling portal', 'pink') + ` appears, and ` + mob.GetCharacterName(true) + ` steps out of it, right before it disappears.`);
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

    if ( lastTipRound == -1 ) {
        return true;
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

    switch( UtilDiceRoll(1, 12) ) {
        case 1:
            if ( charmer.GetStatPoints() > 0 ) {
                mob.Command(`sayto @` + charmer.UserId() + ` It looks like you've got some stat points to spend. Type <ansi fg="command">status train</ansi> to upgrade your stats!`);
            }
            break;
        case 2:
            if ( !charmer.HasQuest(`4-start`) ) {
                mob.Command(`sayto @` + charmer.UserId() + ` There's a guard in the barracks that constantly complains about being hungry. You should <ansi fg="command">ask</ansi> him about it.`);
            }
            break;
        case 3:
            if ( !charmer.HasQuest(`2-start`) ) {
                mob.Command(`sayto @` + charmer.UserId() + ` I have heard the king worries. If we can find an audience with him we can try to <ansi fg="command">ask</ansi> him about a quest. He is north of town square.`);
            }
            break;
        case 4:
            mob.Command(`sayto @` + charmer.UserId() + ` You can find help on many subjects by typing <ansi fg="command">help</ansi>.`);
            break;
        case 5:
            mob.Command(`sayto @` + charmer.UserId() + ` I can create a portal to take us back to <ansi fg="room-title">Town Square</ansi> any time. Just <ansi fg="command">ask</ansi> me about it.`);
            break;
        case 6:
            mob.Command(`sayto @` + charmer.UserId() + ` If you have friends to play with, you can party up! <ansi fg="command">help party</ansi> to learn more.`);
            break;
        case 7:
            mob.Command(`sayto @` + charmer.UserId() + ` You can send a message to everyone using the <ansi fg="command">broadcast</ansi> command.`);
            break
        case 8:
            if ( charmer.GetLevel() < 2 ) {
                mob.Command(`sayto @` + charmer.UserId() + ` There are some <ansi fg="mobname">rats</ansi> to bash around the <ansi fg="room-title">The Sanctuary of the Benevolent Heart</ansi> south of <ansi fg="room-title">Town Square</ansi>. Just don't go TOO far south, it get dangerous!`);
                break;
            }
        case 9:
            if ( charmer.GetLevel() < 2 ) {
                mob.Command(`sayto @` + charmer.UserId() + ` Type <ansi fg="command">status</ansi> to learn about yourself!`);
            }
            break;
        case 10:
            if ( charmer.GetLevel() < 2 ) {
                mob.Command(`sayto @` + charmer.UserId() + ` Killing stuff is a great way to get stronger, but don't pick a fight with the locals!`);
            }
            break;
        default:
            break;
    }

    // Prevent from triggering too often
    mob.SetTempData(`lastTipRound`, roundNow);

    return true;
}


// Things to ask to get a portal created
const homeNouns = ["home", "portal", "return", "townsquare", "town square"];

// Things to ask to shut up the guide
const silenceNouns = ["silence", "quiet", "shut up", "shh"];

const leaveNouns = ["leave", "leave me alone", "die", "quit", "go away", "unfollow", "get lost"];

function onAsk(mob, room, eventDetails) {

    charmedUserId = mob.GetCharmedUserId();

    if ( eventDetails.sourceId != charmedUserId ) {
        return false;
    }

    user = GetUser(eventDetails.sourceId);

    match = UtilFindMatchIn(eventDetails.askText, homeNouns);
    if ( match.found ) {

        if ( user.GetRoomId() == 1 ) {
            mob.Command(`sayto @`+String(eventDetails.sourceId)+` we're already at <ansi fg="room-title">Town Square</ansi>. <ansi fg="command">Look</ansi> around!`);
            return true;
        }

        mob.Command(`sayto @`+String(eventDetails.sourceId)+` back to <ansi fg="room-title">Town Square</ansi>? Sure thing, lets go!`);
        mob.Command(`emote whispers a soft incantation and summons a ` + UtilApplyColorPattern(`glowing portal`, `cyan`) + `.`);

        room.AddTemporaryExit(`glowing portal`, `:cyan`, 0, 3); // roomId zero is start room alias

        return true;
    }

    match = UtilFindMatchIn(eventDetails.askText, silenceNouns);
    if ( match.found ) {

        mob.Command(`sayto @`+String(eventDetails.sourceId)+` I'll try and be quieter.`);
        
        mob.SetTempData(`lastTipRound`, -1);

        return true;
    }


    match = UtilFindMatchIn(eventDetails.askText, leaveNouns);
    if ( match.found ) {

        mob.Command(`sayto @`+String(eventDetails.sourceId)+` I'll be on my way then.`);
        mob.Command(`emote bows and bids you farewell, disappearing into the scenery`);
        mob.Command(`despawn charmed mob expired`)

        return true;
    }

    return false;
}

