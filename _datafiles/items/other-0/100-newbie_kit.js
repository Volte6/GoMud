
const ITEM_LIST = [
  10004, // dagger
  20009, // cloth belt
  20008, // cotton shirt
  20003, // worn boots
  20001, // rat pelt
  20007, // rusty pot
  20006, // tattered pants
  20002, // cape
  20004, // wooden shield
  20005, // copper ring
  30001  // small red potion
];

function onCommand_use(user, item, room) {
    
    
    SendUserMessage(user.UserId(), "You break open the <ansi fg=\"itemname\">"+item.Name()+"</ansi> and loot the contents.");
    SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" breaks open their <ansi fg=\"itemname\">"+item.Name()+"</ansi>, looting the contents.", user.UserId())

    for ( var i=0; i<ITEM_LIST.length; i++) {
        item_id = ITEM_LIST[i];
        itm = CreateItem(item_id);
        SendUserMessage(user.UserId(), "You find a <ansi fg=\"itemname\">"+itm.Name()+"</ansi> inside!");
        user.GiveItem(itm);
    }

    item.AddUsesLeft(-1); // Decrement the uses left by 1
    item.MarkLastUsed(); // Update the last used round number to current
 
    return true;
}
