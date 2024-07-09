
const UnlimitedMinutes = -1;
const SixtyMinutes = 60*60;
const FifteenMinutes = 60*15;
const FiveMinutes = 60*5;

const MOD_SKILL_MIN = 1;
const MOD_SKILL_MAX = 100;

const MOD_LEVELDIFF_MIN = -25;
const MOD_LEVELDIFF_MAX = 25;

const MOD_HEALTHPERCENT_MAX = 50;

const MOD_SIZE_SMALL = 0;
const MOD_SIZE_MEDIUM = -10;
const MOD_SIZE_LARGE = -25;

const FACTOR_IS_AGGRO = .50;


function getTameSkills(sourceActor) {

    allCreatures = sourceActor.GetMiscCharacterDataKeys("tameskill-");

    allTameSkills = [];
    for( var i in allCreatures ) {
        cName = allCreatures[i];
        allTameSkills[cName] = sourceActor.GetMiscCharacterData("tameskill-"+cName);
    }

    return allTameSkills;
}

function modifyTameSkill(sourceActor, targetActor, modifier) {

    targetName = targetActor.GetCharacterName(false);

    allTameSkills = getTameSkills(sourceActor);
    proficiencyModifier = allTameSkills[targetName];
    if ( proficiencyModifier == null ) {
        proficiencyModifier = 0;
    } 
    
    newProficiencyModifier = proficiencyModifier + modifier;
    
    if ( newProficiencyModifier != proficiencyModifier ) {
        if ( newProficiencyModifier < MOD_SKILL_MIN ) {
            newProficiencyModifier = MOD_SKILL_MIN;
        } else if ( newProficiencyModifier > MOD_SKILL_MAX ) {
            newProficiencyModifier = MOD_SKILL_MAX;
        }

        
        targetActor.SetMiscCharacterData("tameskill-"+targetName, newProficiencyModifier);
    }

    return newProficiencyModifier;
}

function calculateChanceIn100(sourceActor, targetActor) {

    targetName = targetActor.GetCharacterName(false);
    
    allTameSkills = getTameSkills(sourceActor);
    proficiencyModifier = allTameSkills[targetName];
    if ( proficiencyModifier == null ) {
        proficiencyModifier = 0;
    } else if ( proficiencyModifier < MOD_SKILL_MIN ) {
        proficiencyModifier = MOD_SKILL_MIN;
    } else if ( proficiencyModifier > MOD_SKILL_MAX ) {
        proficiencyModifier = MOD_SKILL_MAX;
    }

    // Every 10 successes they get better at it.
    proficiencyModifier = Math.ceil( proficiencyModifier / 10 );

    sizeModifier = 0;
    switch( targetActor.GetSize() ) {
        case "large":
            sizeModifier = MOD_SIZE_LARGE;
            break;
        case "medium":
        default:
            sizeModifier = MOD_SIZE_MEDIUM;
        break;
        case "small":
            sizeModifier = MOD_SIZE_SMALL;
        break;
    }
    // console.log('proficiencyModifier: '+proficiencyModifier);

    levelDiff = sourceActor.GetLevel() - targetActor.GetLevel();
    if ( levelDiff > MOD_LEVELDIFF_MAX ) {
        levelDiff = MOD_LEVELDIFF_MAX;
    } else if ( levelDiff < MOD_LEVELDIFF_MIN ) {
        levelDiff = MOD_LEVELDIFF_MIN;
    }

    // console.log('levelDiff: '+levelDiff);

    healthModifier = MOD_HEALTHPERCENT_MAX - Math.ceil( targetActor.GetHealthPct() * MOD_HEALTHPERCENT_MAX );

    // console.log('healthModifier: '+healthModifier);

    aggroModifier = 1;
    if ( targetActor.IsAggro(sourceActor) ) {
        aggroModifier = FACTOR_IS_AGGRO;
    }

    return Math.ceil( (proficiencyModifier + levelDiff + healthModifier + sizeModifier) * aggroModifier );
}

// Called when the casting is initialized (cast command)
// Return false if the casting should be ignored/aborted
function onCast(sourceActor, targetActor) {

    
    if ( targetActor.IsCharmed() ) {
        SendUserMessage(sourceActor.UserId(), 'Already friendly!');
        return false;
    }

    skillLevel = sourceActor.GetSkillLevel("tame");
    charmCt = sourceActor.GetCharmCount();
    if ( charmCt >= skillLevel+1 ) {
        SendUserMessage(sourceActor.UserId(), 'You can only have '+String(skillLevel+1)+' creatures following you at a time.');
        return true;    
    }


    allTameSkills = getTameSkills(sourceActor);
    proficiencyModifier = allTameSkills[targetActor.GetCharacterName(false)];
    if ( proficiencyModifier == null ) {
        SendUserMessage(sourceActor.UserId(), 'You don\'t know how to tame a '+targetActor.GetCharacterName(true)+'.');
        return true;
    }

    chance = calculateChanceIn100(sourceActor, targetActor);
    
    SendUserMessage(sourceActor.UserId(), 'You have a <ansi fg="151">'+chance+'% chance</ansi> to successfully tame the '+targetActor.GetCharacterName(true)+'.');
    SendUserMessage(sourceActor.UserId(), 'You begin to dance in front of the '+targetActor.GetCharacterName(true)+'.');
    SendRoomMessage(sourceActor.GetRoomId(), sourceActor.GetCharacterName(true)+' performs a carefully choreographed dance in front of '+targetActor.GetCharacterName(true)+'.', sourceActor.UserId());

    return true
}

function onWait(sourceActor, targetActor) {

    switch ( UtilDiceRoll(1, 7) ) {
        case 1:
            SendUserMessage(sourceActor.UserId(), 'You make a series of gutteral sounds that seem to distract the '+targetActor.GetCharacterName(true)+'.');
            SendRoomMessage(sourceActor.GetRoomId(), sourceActor.GetCharacterName(true)+' makes a series of gutteral sounds that seem to distract the '+targetActor.GetCharacterName(true)+'.', sourceActor.UserId());
            break;
        case 2:
            SendUserMessage(sourceActor.UserId(), 'You continue to chant in front of the '+targetActor.GetCharacterName(true)+'.');
            SendRoomMessage(sourceActor.GetRoomId(), sourceActor.GetCharacterName(true)+' continues to chant in front of '+targetActor.GetCharacterName(true)+'.', sourceActor.UserId());
            break;
        case 3:
            SendUserMessage(sourceActor.UserId(), 'You fall to the floor and slither like a snake in front of the '+targetActor.GetCharacterName(true)+'.');
            SendRoomMessage(sourceActor.GetRoomId(), sourceActor.GetCharacterName(true)+' falls to the floor and slithers like a snake in front of the '+targetActor.GetCharacterName(true)+'.', sourceActor.UserId());
            break;
        case 4:
            SendUserMessage(sourceActor.UserId(), 'Your body stiffens, and the '+targetActor.GetCharacterName(true)+' becomes alert.');
            SendRoomMessage(sourceActor.GetRoomId(), sourceActor.GetCharacterName(true)+' suddenly stiffens, and the '+targetActor.GetCharacterName(true)+' becomes alert.', sourceActor.UserId());
            break;
        case 5:
            SendUserMessage(sourceActor.UserId(), 'Your run in a circle around the '+targetActor.GetCharacterName(true)+'.');
            SendRoomMessage(sourceActor.GetRoomId(), sourceActor.GetCharacterName(true)+' runs in a circle around the '+targetActor.GetCharacterName(true)+'.', sourceActor.UserId());
            break;
        case 6:
            SendUserMessage(sourceActor.UserId(), 'Your purr ever so gently.');
            SendRoomMessage(sourceActor.GetRoomId(), sourceActor.GetCharacterName(true)+' purrs ever so gently.', sourceActor.UserId());
            break;
        default:
            SendUserMessage(sourceActor.UserId(), 'You whistle several times, changing your pitch ever so slightly.');
            SendRoomMessage(sourceActor.GetRoomId(), sourceActor.GetCharacterName(true)+' whistles several times, changing your pitch ever so slightly.', sourceActor.UserId());
    }

}

// Called when the spell succeeds its cast attempt
// Return true to ignore any auto-retaliation from the target
function onMagic(sourceActor, targetActor) {

    if ( targetActor.IsCharmed() ) {
        SendUserMessage(sourceActor.UserId(), 'Already friendly!');
        return false;
    }

    targetName = targetActor.GetCharacterName(true);
    sourceName = sourceActor.GetCharacterName(true);

    successChance = calculateChanceIn100(sourceActor, targetActor);
    randNumber = UtilDiceRoll(1, 100) - 1;

    if ( randNumber >= successChance ) {
        SendUserMessage(sourceActor.UserId(), 'The '+targetName+' <ansi fg="182">RESISTS</ansi> your attempt to tame it!');
        SendRoomMessage(sourceActor.GetRoomId(), 'The '+targetName+' <ansi fg="182">RESISTS</ansi> '+sourceName+'\'s attempt to tame it!', sourceActor.UserId());

        // modifyTameSkill(sourceActor, targetActor, -1); 
        
        return false;
    }

    modifyTameSkill(sourceActor, targetActor, 1);

    SendUserMessage(sourceActor.UserId(), 'You <ansi fg="151">SUCCESSFULLY</ansi> tame the '+targetName+'!');
    SendRoomMessage(sourceActor.GetRoomId(), sourceName+' <ansi fg="151">SUCCESSFULLY</ansi> tames the '+targetName+'!', sourceActor.UserId());
    
    skillLevel = sourceActor.GetSkillLevel("tame");
    tameRounds = 0;
    switch( skillLevel ) {
        case 4:
            tameRounds = SixtyMinutes;
            break;
        case 3:
            tameRounds = UtilGetSecondsToRounds(SixtyMinutes);
            break;
        case 2:
            tameRounds = UtilGetSecondsToRounds(SixtyMinutes);
            break;
        default:
            tameRounds = UtilGetSecondsToRounds(SixtyMinutes);
    }

    targetActor.CharmSet(sourceActor.UserId(), tameRounds, "emote reverts to a wild state.");

    // Tell the caster about the action
    SendUserMessage(sourceActor.UserId(), 'The '+targetName+' has been tamed by you for '+String(tameRounds)+' rounds!');

    // Tell the room about the heal, except the source and target
    SendRoomMessage(sourceActor.GetRoomId(), sourceName+' tames the '+targetName+'!', sourceActor.UserId(), targetActor.UserId());

 
    return true;
}

