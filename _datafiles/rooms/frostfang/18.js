
const tapestries = ["tapestries", "tapestry", "walls"];
const altar = ["altar"];

function onCommand_look(rest, user, room) {

    matches = UtilFindMatchIn(rest, altar);
    if ( matches.found ) {
        SendUserMessage(user.UserId(), "The altar of the Sanctuary of the Benevolent Heart is an enigmatic sight. Crafted from a dark, veined marble, it stands in stark contrast to the temple's otherwise luminous interior. The altar's edges are adorned with intricate, almost hypnotic patterns that seem to shift and swirl when stared at for too long. At its center, a golden censer continuously emits a fragrant incense. The scent, both sweet and slightly musky, is so captivating that any initial unease evoked by the altar's appearance is quickly replaced by a sense of calm and tranquility, lulling visitors into a state of peaceful oblivion.");
        SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" examines the altar.");   
        SendUserMessage(user.UserId(), "<ansi fg=\"240\">The smell of the insense fills your nostrels, numbing your senses.</ansi>");
        
        user.GiveBuff(2);
        
        return true;
    }

    matches = UtilFindMatchIn(rest, tapestries);
    if ( matches.found ) {
        SendUserMessage(user.UserId(), "The tapestries within the Sanctuary of the Benevolent Heart are vibrant masterpieces, each weaving tales of Frostfang's history and the temple's legacy. From scenes of townsfolk uniting during harsh winters to the legend of a priestess drawing water from a mysterious source. One tapestry stands out with its strangely ominous depiction. It portrays shadowy catacombs, where torches cast eerie glows on ancient bones and forgotten relics. Dyed with natural pigments and crafted with unparalleled skill, these artworks serve as visual scriptures, illustrating religious teachings, historical legends, and mysterious tales.");
        SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" examines the tapestries.");

        return true;
    }

    return false;
}
