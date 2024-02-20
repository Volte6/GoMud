
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

// Called whenever a mob uses the converse command.
function onConverse(message, mob, sourceMob, room) {

}