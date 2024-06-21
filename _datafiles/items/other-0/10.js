

STORY = [
        "",
        "In the ancient days, when the land was known as Dragonsfang, it was a place of",
        "verdant valleys and sunlit peaks. The kingdom thrived in the embrace of ordinary",
        "seasons, with spring's blossom, summer's warmth, autumn's harvest, and winter's",
        "gentle snow. The people of Dragonsfang prospered, the fields yielding bountiful", 
        "crops, their rivers teeming with fish, and their hearts filled with the joy of",
        "life's natural rhythms.",
        "",
        "Dragonsfang was ruled by a wise and benevolent king, whose lineage went back to",
        "the legendary dragon riders. The dragon riders were said to have a bond with the",
        "majestic creatures that soared above the mountains, scales shimmering in the",
        "sunlight. The dragons, guardians of the realm, ensured the kingdom's safety and ",
        "harmony. The people revered these noble beasts, and the bond between dragon and",
        "rider was a sacred trust.",
        "",
        "As centuries passed, the kingdom flourished. Grand castles were built upon the",
        "cliffs, overlooking the lush valleys below. The capital city was a marvel of",
        "architecture, its spires reaching towards the heavens, its streets bustling with",
        "life. Festivals and celebrations marked the changing seasons, each bringing", 
        "a unique charm to the kingdom. The fields to the east and west were fertile and",
        "expansive, feeding the populace and providing trade with neighboring lands.",
        "",
        "However, the tides of time bring change, sometimes swift and unrelenting. One",
        "fateful year, the kingdom began to notice subtle shifts. The winters grew longer",
        "and colder, the snows more relentless. Rivers that once flowed freely began to",
        "freeze earlier each year, and the once mild summers became only memories. The",
        "people of Dragonsfang adapted as best they could, their spirits resilient",
        "against the growing chill.",
        "",
        "Decades turned into centuries, and the changes persisted. The dragons, once a",
        "common sight in the skies, became rarer. Their absence was keenly felt,",
        "their protective presence a mere echo of the past. The dragon riders, too, faded",
        "into legend, their stories told in hushed tones by the fireside. The once grand",
        "capital, now known as Frostfang, bore the weight of unending winter.",
        "",
        "The kingdom's architecture evolved to withstand the harsh climate. Houses were",
        "built sturdier, their walls thicker, and their roofs designed to bear the weight",
        "of constant snow. The people learned to cultivate the hardy crops that could",
        "survive the frigid temperatures, and the livestock adapted to the colder",
        "environment. The fields to the east and west, though now perpetually blanketed",
        "in snow, still supported life through the perseverance of the Frostfang people.",
        ""
        ];


function onCommand_read(user, item, room) {

    console.log("LEN", String(STORY.length));

    SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" thumbs through their <ansi fg=\"item\">"+item.Name()+"</ansi> book.", user.UserId());   

    SendUserMessage(user.UserId(), "");
    SendUserMessage(user.UserId(), "<ansi fg=\"14\">The History of Frostfang</ansi>");
    
    for (var i=0; i<STORY.length; i++) {
        SendUserMessage(user.UserId(), "<ansi fg=\"3\">"+STORY[i]+"</ansi>");
    }

    return true;
}

