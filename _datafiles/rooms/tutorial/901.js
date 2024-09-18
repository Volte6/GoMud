

const allowed_commands = ["help", "broadcast", "look"];
const teach_commands = ["status", "inventory", "experience", "conditions", "south"];
const teacherMobId = 57;
const teacherName = "Orb of Reflection";
const firstItemId = 10001;

var commandNow = 0; // Which command they are on
var canGoSouth = false;




// Generic Command Handler
function onCommand(cmd, rest, user, room) {

    ignoreCommand = false;

    teacherMob = getTeacher(room);

    // Make sure they are only doing stuff that's allowed.

    if ( cmd == "south" && !canGoSouth ) {
        teacherMob.Command("say Not so hasty! Lets finish up here before you leave the area.");
        ignoreCommand = true;
    }

    if ( rest.length > 0 ) {
        cmd += ' ' + rest;
    }

    if ( teach_commands[commandNow] == cmd ) {
        
        teacherMob.Command("say Good job!", 4);

        if ( cmd == "status" ) {
            teacherMob.Command('say You can see how much gold you carry, your Level, and even attributes like Strength and Smarts.', 4);
            teacherMob.Command('say It\'s a lot of information, but you quickly learn to only pay attention to the important stuff.', 4);
        }

        if ( cmd == "inventory" ) {
            teacherMob.Command('say Hmm, it doesn\'t look like you\'re carrying much other than that sharp stick.', 4);
            teacherMob.Command('say Remember, you can <ansi fg="command">look</ansi> at stuff you\'re carrying any time you want.', 4);
        }

        commandNow++;
    } else {

        if ( allowed_commands.includes(cmd) || teach_commands.slice(0, commandNow).includes(cmd) ) {
            return false;
        }
        
        ignoreCommand = true;
    }

    switch (commandNow) {
        case 0:

            if ( !user.HasItemId(firstItemId) ) {
                itm = CreateItem(firstItemId);
                user.GiveItem(itm);
            }
            
            teacherMob.Command('say To see all of your characters stats, type <ansi fg="command">status</ansi>.', 4);
            break;
        case 1:
            teacherMob.Command('say To only peek at your inventory, type <ansi fg="command">inventory</ansi>.', 4);
            break;
        case 2:
            teacherMob.Command('say As you solve quests and defeat enemies in combat, you\'ll gain experience points and your character will "Level up".', 4);
            teacherMob.Command('say For quick look at your progress, type <ansi fg="command">experience</ansi>.', 4);
            break;
        case 3:
            teacherMob.Command('emote touches you and you feel more focused.', 4);
            user.GiveBuff(32);
            teacherMob.Command('say Sometimes you might become afflicted with a condition. Conditions can have good or bad effects.', 4);
            teacherMob.Command('say type <ansi fg="command">conditions</ansi> to see any statuses affecting you.', 4);
            break;
        case 4:
            user.GiveBuff(-32);
            teacherMob.Command('say head <ansi fg="command">south</ansi> for the next lesson.', 4);
            canGoSouth = true;
        default:
            room.SetLocked("south", false);
            break;
    }
    
    return ignoreCommand;
}




// If there is no book here, add the book item
function onEnter(user, room) {
    room.SetLocked("west", true);
    
    teacherMob = getTeacher(room);
    teacherMob.Command('say Hi! I\'m here to teach you about inspecting your characters information.', 4);
    teacherMob.Command('say To get a detailed view of a LOT of information all at once, type <ansi fg="command">status</ansi> and hit enter.', 4);
}



function onExit(user , room) {
    // Destroy the guide (cleanup)
    destroyTeacher(room);
    canGoSouth = false;
    commandNow = 0;
}



function onLoad(room) {
    getTeacher(room);
    canGoSouth = false;
    commandNow = 0;
}


function getTeacher(room) {

    var mobActor = null;

    mobIds = room.GetMobs();
    
    for ( i in mobIds ) {
        mobActor = GetMob(mobIds[i]);
        if ( mobActor.MobTypeId() == teacherMobId ) {
            return mobActor;
        }
    }

    mobActor = room.SpawnMob(teacherMobId);
    mobActor.SetCharacterName(teacherName);

    return mobActor;
}
function destroyTeacher(room) {

    var mobActor = null;

    mobIds = room.GetMobs();
    
    for ( i in mobIds ) {
        mobActor = GetMob(mobIds[i]);
        if ( mobActor.MobTypeId() == teacherMobId ) {
            mobActor.Command(`suicide vanish`);
        }
    }
}