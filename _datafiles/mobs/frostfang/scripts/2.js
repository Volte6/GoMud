
// Return false to negate the damage and messages
function onHurt(mob, room, eventDetails) {
    console.log("onHurt()");
    console.log(mob);
    console.log(room);
    console.log(eventDetails);
    mob.Command("say hey you hurt me, "+String(eventDetails.sourceId)+ ". You did "+String(eventDetails.damage)+" damage to me.");
    console.log("onHurt() DONE");
}

function onDie(mob, room, eventDetails) {
    console.log("onDie()");
    console.log(mob);
    console.log(room);
    console.log(eventDetails);
    
    mob.Command("say im dead");
    SendRoomMessage(room.RoomId(), "User " + String(eventDetails.sourceId) + " killed " + mob.GetCharacterName() + "!");
    console.log("onDie() DONE");
}

function onIdle(mob, room) {

    var random = Math.floor(Math.random() * 8);
    switch (random) {
        case 0:
            mob.Command("emote flexes his muscles");
            return true;
        case 1:
            return true; // does nothing.
        case 2:
        case 3:
            mob.Command("wander");
            return true;
        default:
            break;
    }

    return false;
}