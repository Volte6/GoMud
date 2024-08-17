
const nouns = ["train", "training", "forest", "mushroom", "mushrooms", "animals", "plants"]


function onAsk(mob, room, eventDetails) {

    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }

    match = UtilFindMatchIn(eventDetails.askText, nouns);
    if ( match.found ) {

        mob.Command("emote sighs.")
        mob.Command("say It's so sad, trying to restore the forest to its old self.")
        mob.Command("say My beautiful mushrooms are slowly disappearing.")

        return true;
    }

    return false;
}

function onGive(mob, room, eventDetails) {

    if (eventDetails.sourceType == "mob") {
        return false;
    }

    if ( eventDetails.gold > 0 ) {
        mob.Command("say I have no use for this.");
        mob.Command("drop "+String(eventDetails.gold)+" gold");
        return true;
    }

    if (eventDetails.item) {
        if (eventDetails.item.ItemId != 30007) {
            mob.Command("look !"+String(eventDetails.item.ItemId));
            mob.Command("say I have no use for this.");
            mob.Command("drop !"+String(eventDetails.item.ItemId), UtilGetSecondsToTurns(5));
            return true;
        }
    }

    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }

    sayThankYou = "Thank you! Each piece of mushroom can be used to grow another mushroom!";
    emoteDraw = "draws a strange symbol into the air.";
    mobName = mob.GetCharacterName(true);

    mob.Command("say "+sayThankYou);
    mob.Command("emote "+emoteDraw);

    // Mimick the say and emote because by the time they are sent the user has moved rooms.
    SendUserMessage(eventDetails.sourceId, mobName + " says, \"<ansi fg=\"yellow\">"+sayThankYou+"</ansi>\"");
    SendUserMessage(eventDetails.sourceId, mobName + " <ansi fg=\"blue\">draws a strange symbol into the air</ansi>");

    SendUserMessage(eventDetails.sourceId, 'Suddenly, you are sucked into a strange portal.');
    SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" is sucked into a strange portal.", user.UserId());

    SendRoomMessage(830, user.GetCharacterName(true)+" appears through a portal.");
       
    user.MoveRoom(830);

    return true;
}

