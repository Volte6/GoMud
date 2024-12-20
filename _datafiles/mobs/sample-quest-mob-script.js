

// Things to ask the mob to initate the quest
const questStartSubjects    = ["quest", "desire", "help"];

// Things to ask the mob once the quest is started, seeking more information
const questInfoSubjects     = ["where", "how"];

// The item this mob wants in order to complete the quest.
const REQUIRED_ITEM_ID      = 10001;

// The gold this mob wants in order to complete the quest
const REQUIRED_GOLD_AMOUNT  = 10;

// This corresponds to the quest defined in the _datafiles/quests/ folder.
const QUEST_START_ID        = "1000000-start"       // All quests begin with #-start
const QUEST_NEXT_STEP_ID    = "1000000-givegold"    // Quest steps can be called #-anything
const QUEST_END_ID          = "1000000-end"         // All quests end with #-end


//
// The onAsk() function handles when players invoke the `ask` command
//
function onAsk(mob, room, eventDetails) {

    //
    // Get the user, and if the user is invalid, skip the script (something is wrong)
    //
    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }

    //
    // Do they have the quest end id? Then they have already completed the quest.
    //
    if ( user.HasQuest(QUEST_END_ID) ) {
        mob.Command("say Your help is no longer needed.");
        return true;
    }

    //
    // Do this part of they have not started the quest yet
    //
    if ( !user.HasQuest(QUEST_START_ID) ) {

        //
        // Search the text they inputted for the "ask" command for one of the questStartSubjects
        //
        match = UtilFindMatchIn(eventDetails.askText, questStartSubjects);

        if ( match.found ) {

            mob.Command("emote smiles.");
            mob.Command("say I would really like a sharp stick!");
            //
            // Give them the start quest id
            //
            user.GiveQuest(QUEST_START_ID);

        }

        return true;
    }

    //
    // By this point in the script we know they've at least started the quest
    // Lets see if they are asking any follow up questions for more info.
    //
    match = UtilFindMatchIn(eventDetails.askText, questInfoSubjects);
    if ( match.found ) {
        mob.Command("emote thinks hard for a moment.");
        mob.Command("say You can get sharp sticks from a shop, and gold from selling objects, or possibly killing bad guys and looting them.");
        return true;
    }
    
    return true;
}

//
// The onGive() function is invoked when players give the mob an item or gold.
//
function onGive(mob, room, eventDetails) {

    //
    // Get the user, and if the user is invalid, skip the script (something is wrong)
    //
    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }
    
    
    //
    // Did they give an item?
    //
    if ( eventDetails.item.ItemId ) {

        console.log("ITEMID", eventDetails.item.ItemId)
        
        //
        // If the item they gave isn't the desired item id, give it back.
        //
        if (eventDetails.item.ItemId != REQUIRED_ITEM_ID) {
            
            mob.Command("say That's very kind, but I don't need this right now.");

            // Use special shorthand such as !{item_id} to avoid accidentally dropping the wrong item by name
            // See: internal/scripting/README.md for other shorthand examples.
            // user object actually has a ShorthandId() function to make that easy.
            mob.Command("give !" + String(eventDetails.item.ItemId) + " " + user.ShorthandId()); // Give it to the player using shorthand

            return true;
        }


        //
        // If they've already done this step of the quest (and have the next step quest token), reject the offering.
        //
        if ( user.HasQuest(QUEST_NEXT_STEP_ID) ) {
            mob.Command("say I already have the stick you gave me. I don't need another.");
            mob.Command("give !" + String(eventDetails.item.ItemId) + " " + user.ShorthandId()); // Give it to the player using shorthand
            return true;
        }

        //
        // By this point in the script, we know it's the right item id.
        //

        mob.Command("say Thank you so much! That's the perfect stick!");
        mob.Command("say Do you think you could spare 10 gold?");

        //
        // Give them the next step of the quest
        //
        user.GiveQuest(QUEST_NEXT_STEP_ID)

        return true;
    }


    //
    // Did they give gold?
    //
    if ( eventDetails.gold > 0 ) {

        //
        // If they don't have the part of the quest that asks for gold yet
        // Reject their offering for now.
        //
        if ( !user.HasQuest(QUEST_NEXT_STEP_ID) ) {
            mob.Command("say We aren't quite there yet.");
            mob.Command("give "+String(eventDetails.gold)+" gold " + user.ShorthandId()); // Give it to the player using shorthand
            return true;
        }

        //
        // If they gave less than REQUIRED_GOLD_AMOUNT, reject it
        //
        if ( eventDetails.gold < REQUIRED_GOLD_AMOUNT ) {
            mob.Command("say We aren't quite there yet.");
            mob.Command("give "+String(eventDetails.gold)+" gold " + user.ShorthandId()); // Give it to the player using shorthand
            return true;
        }


        //
        // If they gave less than REQUIRED_GOLD_AMOUNT, reject it
        //
        if ( eventDetails.gold >= REQUIRED_GOLD_AMOUNT ) {
            
            mob.Command("say Great thanks!");

            //
            // If they gave too much gold, lets give them back the change.
            //
            excessGold = eventDetails.gold - REQUIRED_GOLD_AMOUNT;
            if ( excessGold > 0 ) {
                mob.Command("say Here's your change.")
                mob.Command("give "+String(excessGold)+" gold " + user.ShorthandId()); // Give it to the player using shorthand
            }

            //
            // They have now completed the entire quest, all steps are complete.
            //
            user.GiveQuest(QUEST_END_ID)

            return true;
        }

        return true;
    }

    return false;
}


//
// The onIdle() function is invoked whenver the mob is sitting around bored.
// The frequency is dictated by the mob's "activitylevel" defined in its flatfile.
//
function onIdle(mob, room) {

    // 25% chance of saying something in a given round
    if ( UtilGetRoundNumber() % 4 == 0 ) {
        
        mob.Command('say I have a quest for a new adventurer. You can <ansi fg="command">ask</ansi> me about it.');
        
        // Returning true from this function says we handled the idle, so do not try to do other random idle behaviors.
        return true;
    }

    //
    // otherwise return false, allow the mob to perform idlecommands it may have defined.
    //
    return false;
}