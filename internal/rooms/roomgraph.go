package rooms

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/volte6/gomud/internal/users"
)

type roomNode struct {
	RoomId      int
	Symbol      rune
	Legend      string
	Exits       map[string]*roomNode
	SecretExits map[string]struct{} // Just a flag for whether an exit key is secret
	xPos        int                 // Its x position relative to the root node
	yPos        int                 // Its y position relative to the root node
	Sprawl      int                 // how far from the start point this is
	MobIds      []int               // all mob instance ids in this room
	UserIds     []int               // all user ids in this room
}

type symbolOverride struct {
	Symbol rune
	Legend string
}

type RoomGraph struct {
	Root            *roomNode
	trackedRoomIds  map[int]*roomNode      // Shortcut to jump to any room id
	minX            int                    // The smallest tracked x position
	maxX            int                    // The largest tracked x position
	minY            int                    // The smallest tracked y position
	maxY            int                    // The largest tracked y position
	width           int                    // The width of the 2D map
	height          int                    // The height of the 2D map
	limitWidth      int                    // The maximum width at which we should stop crawling
	limitHeight     int                    // The maximum height at which we should stop crawling
	povUserId       int                    // The user id that is the target of the map. Zero if admin or we want a fuller map
	mode            MapMode                // Flags for special rendering behaviors
	maxSprawl       int                    // The maximum node distance to search
	roomLimits      map[int]struct{}       // An optional list of room ids that the map should be restricted to (if any).
	forceRoomSymbol map[int]symbolOverride // An optional list of room ids that the map should be restricted to (if any).
}
type foundRoomExits struct {
	roomNode *roomNode
	exits    map[string]RoomExit
}

type directionDelta struct {
	Dx    int
	Dy    int
	Dz    int
	Arrow rune
}

// A simplified 2D map of room id's and their exits
type Map2DNode struct {
	RoomId      int
	Symbol      rune
	Legend      string // The name for this room on the legend (if needed)
	Exits       []directionDelta
	SecretExits []directionDelta
}
type Map1D []*Map2DNode
type Map2D []Map1D

var (
	DirectionDeltas = map[string]directionDelta{
		"north":     {0, -1, 0, '│'},
		"south":     {0, 1, 0, '│'},
		"west":      {-1, 0, 0, '─'},
		"east":      {1, 0, 0, '─'},
		"northwest": {-1, -1, 0, '╲'},
		"northeast": {1, -1, 0, '╱'},
		"southwest": {-1, 1, 0, '╱'},
		"southeast": {1, 1, 0, '╲'},

		// Double spaced away
		"north-x2":     {0, -2, 0, '│'},
		"south-x2":     {0, 2, 0, '│'},
		"west-x2":      {-2, 0, 0, '─'},
		"east-x2":      {2, 0, 0, '─'},
		"northwest-x2": {-2, -2, 0, '╲'},
		"northeast-x2": {2, -2, 0, '╱'},
		"southwest-x2": {-2, 2, 0, '╱'},
		"southeast-x2": {2, 2, 0, '╲'},
		// Double spaced away
		"north-x3":     {0, -3, 0, '|'},
		"south-x3":     {0, 3, 0, '|'},
		"west-x3":      {-3, 0, 0, '─'},
		"east-x3":      {3, 0, 0, '─'},
		"northwest-x3": {-3, -3, 0, '╲'},
		"northeast-x3": {3, -3, 0, '╱'},
		"southwest-x3": {-3, 3, 0, '╱'},
		"southeast-x3": {3, 3, 0, '╲'},
		// The following are for rendering exits that are gaps in the map
		"north-gap":     {0, -1, 0, ' '},
		"south-gap":     {0, 1, 0, ' '},
		"west-gap":      {-1, 0, 0, ' '},
		"east-gap":      {1, 0, 0, ' '},
		"northwest-gap": {-1, -1, 0, ' '},
		"northeast-gap": {1, -1, 0, ' '},
		"southwest-gap": {-1, 1, 0, ' '},
		"southeast-gap": {1, 1, 0, ' '},
		// 2 space gap
		"north-gap2":     {0, -2, 0, ' '},
		"south-gap2":     {0, 2, 0, ' '},
		"west-gap2":      {-2, 0, 0, ' '},
		"east-gap2":      {2, 0, 0, ' '},
		"northwest-gap2": {-2, -2, 0, ' '},
		"northeast-gap2": {2, -2, 0, ' '},
		"southwest-gap2": {-2, 2, 0, ' '},
		"southeast-gap2": {2, 2, 0, ' '},
		// 3 space gap
		"north-gap3":     {0, -3, 0, ' '},
		"south-gap3":     {0, 3, 0, ' '},
		"west-gap3":      {-3, 0, 0, ' '},
		"east-gap3":      {3, 0, 0, ' '},
		"northwest-gap3": {-3, -3, 0, ' '},
		"northeast-gap3": {3, -3, 0, ' '},
		"southwest-gap3": {-3, 3, 0, ' '},
		"southeast-gap3": {3, 3, 0, ' '},
	}
)

type MapMode uint8

const (

	//
	// Additionally maps may have a maximum distance they will map out to.
	// Limitaitons (if any) of how much map to render
	MapLimitNone   int = 0
	MapLimitLevel1 int = 4
	MapLimitLevel2 int = 6
	MapLimitLevel3 int = 8

	//
	// The mode imposes limitations on what the graph can connect
	// Different ways mapping can function:
	// 0. Map everything within a certain number of rooms of the character
	//    - Useful for a "bought" map of the area
	// 1. Only map rooms that a userId has visited in the last X minutes (room.HasVisited(userId))
	//    - Useful for "tracking" other players
	// 2. Only map rooms that the userId remembers visiting (room history stack)
	//    - Useful for a basic mapping skill
	MapModeAll           MapMode = 0
	MapModeAllButSecrets MapMode = 1
	MapModeTracking      MapMode = 2
	MapModeRadius        MapMode = 4
)

func NewRoomGraph(maxWidth int, maxHeight int, uidPov int, mode MapMode) *RoomGraph {
	rGraph := &RoomGraph{
		width:           1,
		height:          1,
		limitWidth:      maxWidth,
		limitHeight:     maxHeight,
		trackedRoomIds:  make(map[int]*roomNode),
		povUserId:       uidPov,
		mode:            mode,
		roomLimits:      make(map[int]struct{}),
		forceRoomSymbol: make(map[int]symbolOverride),
	}

	if mode == MapModeRadius {
		if user := users.GetByUserId(uidPov); user != nil {
			rGraph.maxSprawl = user.Character.GetMapSprawlCapacity()
		}
	}

	return rGraph
}

func newNode(roomId int, roomSymbol string, roomLegend string, allMobsInstanceIds []int, allPlayerUserIds []int) *roomNode {
	if len(roomSymbol) == 0 {
		roomSymbol = defaultMapSymbol // *
	}
	if len(roomLegend) == 0 {
		roomLegend = "Room"
	}
	return &roomNode{
		RoomId:      roomId,
		Exits:       make(map[string]*roomNode),
		SecretExits: make(map[string]struct{}),
		Symbol:      []rune(roomSymbol)[0],
		Legend:      roomLegend,
		MobIds:      allMobsInstanceIds,
		UserIds:     allPlayerUserIds,
	}
}

func (m *Map2D) Strings() []string {
	var output []string = make([]string, len(*m))
	for _, row := range *m {
		output = append(output, row.String())
	}
	return output
}

func (m *Map2D) String() string {
	var output strings.Builder
	for _, row := range *m {
		output.WriteString(row.String())
		output.WriteString("\n")
	}
	return output.String()
}

func (m *Map1D) String() string {
	var output strings.Builder
	for _, nodeInfo := range *m {
		if nodeInfo == nil {
			output.WriteString("     |")
			continue
		}
		output.WriteString(fmt.Sprintf("%-4d%s|", nodeInfo.RoomId, string(nodeInfo.Symbol)))
	}
	return output.String()
}

func (r *RoomGraph) RoomCount() int {
	return len(r.trackedRoomIds)
}

func (r *RoomGraph) RoomIds() []int {
	allRoomIds := make([]int, 0, len(r.trackedRoomIds))
	for roomId, _ := range r.trackedRoomIds {
		allRoomIds = append(allRoomIds, roomId)
	}
	return allRoomIds
}

func (r *RoomGraph) RoomIdsWithPlayers() []int {
	allRoomIds := make([]int, 0, len(r.trackedRoomIds))
	for roomId, roomNode := range r.trackedRoomIds {
		if len(roomNode.UserIds) > 0 {
			allRoomIds = append(allRoomIds, roomId)
		}
	}
	return allRoomIds
}

func (r *RoomGraph) RoomIdsWithMobs() []int {
	allRoomIds := make([]int, 0, len(r.trackedRoomIds))
	for roomId, roomNode := range r.trackedRoomIds {
		if len(roomNode.MobIds) > 0 {
			allRoomIds = append(allRoomIds, roomId)
		}
	}
	return allRoomIds
}

// Whatever the room normally shows, it will show this instead.
func (r *RoomGraph) AddRoomSymbolOverrides(symbol rune, legend string, roomIds ...int) {
	for _, roomId := range roomIds {
		r.forceRoomSymbol[roomId] = symbolOverride{
			Symbol: symbol,
			Legend: legend,
		}
	}
}

// Returns a 2D slice of rooms id's, and the coordinates fo the center room within the 2D slice
// If zero (0) is provided for mapMaxWidth and mapMaxHeight,
// it will default to the width and height of the graph.
func (r *RoomGraph) Generate2DMap(mapMaxWidth int, mapMaxHeight int, centerRoomId int) (map2DResult Map2D, centerRoomX int, centerRoomY int) {

	centerRoom := r.findRoom(centerRoomId)
	if centerRoom == nil {
		map2DResult = make(Map2D, 1)
		map2DResult[0] = make(Map1D, 1)
		return map2DResult, 0, 0
	}

	var xStart, yStart, xEnd, yEnd int

	// If zero is provided, grab the entire width and move the center point to the new center
	if mapMaxWidth == 0 && mapMaxHeight == 0 {
		mapMaxWidth, mapMaxHeight = r.width, r.height
		xStart, xEnd = r.minX, r.minX+mapMaxWidth
		yStart, yEnd = r.minY, r.minY+mapMaxHeight
	} else {
		xStart = centerRoom.xPos - (mapMaxWidth >> 1)
		if xStart < r.minX {
			xStart = r.minX
		}
		xEnd = xStart + mapMaxWidth

		yStart = centerRoom.yPos - (mapMaxHeight >> 1)
		if yStart < r.minY {
			yStart = r.minY
		}
		yEnd = yStart + mapMaxHeight
	}

	xLeftPad := (mapMaxWidth - r.width) >> 1
	if xLeftPad < 0 {
		xLeftPad = 0
	}

	yTopPad := (mapMaxHeight - r.height) >> 1
	if yTopPad < 0 {
		yTopPad = 0
	}

	// Create the 2D map
	map2DResult = make(Map2D, mapMaxHeight)
	for i := range map2DResult {
		map2DResult[i] = make(Map1D, mapMaxWidth)
	}

	// Every roomNode has an x╱y position relative to the root node.
	// To ensure that every room is above 0,0, we need to offset the
	// x╱y position by the minimum x╱y position.
	// Then write the node's room id to the 2D map at that position.
	for _, roomNode := range r.trackedRoomIds {

		if r.maxSprawl > 0 && roomNode.Sprawl > r.maxSprawl {
			continue
		}
		if boundaryCheck(roomNode.xPos, roomNode.yPos, xStart, xEnd, yStart, yEnd) {

			symbol := roomNode.Symbol
			legend := roomNode.Legend

			if symbolInfo, ok := r.forceRoomSymbol[roomNode.RoomId]; ok {
				symbol = symbolInfo.Symbol
				legend = symbolInfo.Legend
			}

			node2D := &Map2DNode{
				RoomId:      roomNode.RoomId,
				Exits:       make([]directionDelta, 0, 3),
				SecretExits: make([]directionDelta, 0),
				Symbol:      symbol,
				Legend:      legend,
			}

			// Mark the center point.
			//if roomNode.RoomId == centerRoomId {
			//	node2D.Symbol = '@'
			//}

			for exitDirection, _ := range roomNode.Exits {
				if directionDelta, ok := DirectionDeltas[exitDirection]; ok {
					node2D.Exits = append(node2D.Exits, directionDelta)
				}
			}

			for exitDirection, _ := range roomNode.SecretExits {
				if directionDelta, ok := DirectionDeltas[exitDirection]; ok {
					node2D.SecretExits = append(node2D.SecretExits, directionDelta)
				}
			}

			/*
				for directionName, directionDelta := range DirectionDeltas {
					if _, ok := roomNode.Exits[directionName]; ok {
						node2D.Exits = append(node2D.Exits, directionDelta)
					}

					if _, ok := roomNode.SecretExits[directionName]; ok {
						node2D.SecretExits = append(node2D.SecretExits, directionDelta)
					}
				}
			*/
			map2DResult[roomNode.yPos-yStart+yTopPad][roomNode.xPos-xStart+xLeftPad] = node2D
		}
	}

	centerRoomX = centerRoom.xPos - xStart + xLeftPad
	centerRoomY = centerRoom.yPos - yStart + yTopPad

	return map2DResult, centerRoomX, centerRoomY
}

func (r *RoomGraph) findRoom(roomId int) *roomNode {
	if foundRoom, ok := r.trackedRoomIds[roomId]; ok {
		return foundRoom
	}
	return nil
}

func (r *RoomGraph) addNode(sourceRoomNode *roomNode, direction string, roomId int, isSecretExit bool) *foundRoomExits {

	// Add a new room exit to an existing graph Node.
	// Return a list of new exits to add to the new room

	// 1. Check if the room has already been added to the graph
	// 2. If it has, just add the exit to the existing room
	//    It is presumed that the exits have already been addressed
	// 3. If it has not, add the new room to the graph and add the exit to the new room linked to the new room node
	//   Then return a list of new exits found in the new room, so that they can be added to the graph
	//  This is to avoid a depth-first search, and instead stay breadth-first
	// Get the target room data
	newRoomData := LoadRoom(roomId)

	if newRoomData == nil {
		return nil
	}

	// usreIdPov 0 has no limitations
	if r.mode == MapModeAllButSecrets {
		if isSecretExit {
			return nil
		}
	}

	if r.povUserId > 0 {
		// Secret exits aren't included unless the player has recently been there
		if isSecretExit {
			if !newRoomData.HasVisited(r.povUserId, VisitorUser) {
				return nil
			}
		} else {

			if r.mode == MapModeTracking {
				// If the player isn't around or they haven't visited the room, recently
				if !newRoomData.HasVisited(r.povUserId, VisitorUser) {
					return nil
				}
			}

		}
	}

	// If this room has already been added, don't search its exits
	existingNewRoom := r.trackedRoomIds[newRoomData.RoomId]
	if existingNewRoom != nil {
		// We just want to connect source room exit to this existing target node
		if isSecretExit {
			sourceRoomNode.SecretExits[direction] = struct{}{}
		}
		sourceRoomNode.Exits[direction] = existingNewRoom
		// Don't crawl into this room again, just return
		return nil
	}

	// This is a new graph node

	mapSymbol := newRoomData.MapSymbol
	mapLegend := newRoomData.MapLegend
	if mapSymbol == `` {
		b := newRoomData.GetBiome()
		if b.symbol != 0 {
			mapSymbol = string(b.symbol)
		}
		if b.name != `` {
			mapLegend = b.name
		}
	}

	// This is a new room so we need to handle it accordingly.
	newRoomNode := newNode(newRoomData.RoomId, mapSymbol, mapLegend, newRoomData.GetMobs(), newRoomData.GetPlayers())
	newRoomNode.Sprawl = sourceRoomNode.Sprawl + 1

	// Track the position relative to the source room, for the new room.
	if exitDelta, ok := DirectionDeltas[direction]; ok {
		newRoomNode.xPos += sourceRoomNode.xPos + exitDelta.Dx
		newRoomNode.yPos += sourceRoomNode.yPos + exitDelta.Dy
	}

	// Mark it as tracked so that we don't recurse into it again
	r.trackedRoomIds[newRoomNode.RoomId] = newRoomNode

	// Now add it to the source rooms exits
	sourceRoomNode.Exits[direction] = newRoomNode

	// Was the source room's exit to this new room a secret? If so mark it as such.
	if isSecretExit {
		sourceRoomNode.SecretExits[direction] = struct{}{}
	}

	// Each time we add a new node, we have to check if it expands the boundaries of the graph
	// Expand the tracked boundaries if needed
	if newRoomNode.xPos < r.minX {
		r.minX = newRoomNode.xPos
	} else if newRoomNode.xPos > r.maxX {
		r.maxX = newRoomNode.xPos
	}
	if newRoomNode.yPos < r.minY {
		r.minY = newRoomNode.yPos
	} else if newRoomNode.yPos > r.maxY {
		r.maxY = newRoomNode.yPos
	}

	// Calculate╱set the new width╱height
	r.width = r.maxX - r.minX + 1
	r.height = r.maxY - r.minY + 1

	// Finally, we crawl the new room's exit data
	// If they don't sit outside the boundaries of the graph, we add them to the returned stack items
	newRoomsToAdd := make(map[string]RoomExit)

	// Track whether we have hit the limit
	graphMaxedWidth := r.width > r.limitWidth
	graphMaxedHeight := r.height > r.limitHeight

	for realExitDirection, exitInfo := range newRoomData.Exits {
		exitDirection := realExitDirection
		if exitInfo.MapDirection != `` {
			exitDirection = exitInfo.MapDirection
		}
		if dInfo, ok := DirectionDeltas[exitDirection]; ok {
			// Make sure it stays within boundaries
			if graphMaxedWidth {
				if dInfo.Dx+newRoomNode.xPos < r.minX || dInfo.Dx+newRoomNode.xPos > r.maxX { // If it would stretch the X boundaries, skip it
					continue
				}
			}
			if graphMaxedHeight {
				if dInfo.Dy+newRoomNode.yPos < r.minY || dInfo.Dy+newRoomNode.yPos > r.maxY { // If it would stretch the Y boundaries, skip it
					continue
				}
			}

			// Contains .RoomId and .Secret
			newRoomsToAdd[exitDirection] = newRoomData.Exits[realExitDirection]
		}
	}

	/*
		for directionName, dInfo := range DirectionDeltas {

			// If an exit exists in this direction, look into it
			if _, ok := newRoomData.Exits[directionName]; ok {

				// Make sure it stays within boundaries
				if graphMaxedWidth {
					if dInfo.Dx+newRoomNode.xPos < r.minX || dInfo.Dx+newRoomNode.xPos > r.maxX { // If it would stretch the X boundaries, skip it
						continue
					}
				}
				if graphMaxedHeight {
					if dInfo.Dy+newRoomNode.yPos < r.minY || dInfo.Dy+newRoomNode.yPos > r.maxY { // If it would stretch the Y boundaries, skip it
						continue
					}
				}

				// Contains .RoomId and .Secret
				newRoomsToAdd[directionName] = newRoomData.Exits[directionName]

			}
		}
	*/

	return &foundRoomExits{
		roomNode: newRoomNode,
		exits:    newRoomsToAdd,
	}

}

func (r *RoomGraph) Build(rootRoomId int, overrideRoomIdSymbols map[int]rune) error {

	if r.Root != nil {
		return nil
		//return errors.New("this rooom graph has already started (or completed) building")
	}

	roomNow := LoadRoom(rootRoomId)
	if roomNow == nil {
		return fmt.Errorf("could not load room id %d", rootRoomId)
	}

	mapSymbol := roomNow.MapSymbol
	mapLegend := roomNow.MapLegend
	if mapSymbol == `` {
		b := roomNow.GetBiome()
		if b.symbol != 0 {
			mapSymbol = string(b.symbol)
		}
		if b.name != `` {
			mapLegend = b.name
		}
	}

	// Create the root node
	var newRoomNode *roomNode = newNode(roomNow.RoomId, mapSymbol, mapLegend, roomNow.GetMobs(), roomNow.GetPlayers())
	r.Root = newRoomNode                          // Make it the root.
	r.trackedRoomIds[r.Root.RoomId] = newRoomNode // Mark it tracked

	roomStack := make([]*foundRoomExits, 0, 100)

	// Now start the crawl by adding exits.
	for directionName, exitInfo := range roomNow.Exits {

		if exitInfo.MapDirection != `` {
			directionName = exitInfo.MapDirection
		}

		if _, ok := DirectionDeltas[directionName]; ok {
			if addlExits := r.addNode(r.Root, directionName, exitInfo.RoomId, exitInfo.Secret); addlExits != nil {
				roomStack = append(roomStack, addlExits)
			}
		}
	}
	/*
		for directionName, _ := range DirectionDeltas {
			if exitInfo, ok := roomNow.Exits[directionName]; ok {
				if addlExits := r.addNode(r.Root, directionName, exitInfo.RoomId, exitInfo.Secret); addlExits != nil {
					roomStack = append(roomStack, addlExits)
				}
			}
		}
	*/

	for len(roomStack) > 0 {

		nextRoomExit := roomStack[0] // Grab the first room in the stack
		roomStack = roomStack[1:]    // Remove it from the stack

		// Possibly override the symbol for this room
		if overrideRoomIdSymbols != nil && overrideRoomIdSymbols[nextRoomExit.roomNode.RoomId] != 0 {
			nextRoomExit.roomNode.Symbol = overrideRoomIdSymbols[nextRoomExit.roomNode.RoomId]
		}

		for directionName, roomExit := range nextRoomExit.exits {
			if addlExits := r.addNode(nextRoomExit.roomNode, directionName, roomExit.RoomId, roomExit.Secret); addlExits != nil {
				roomStack = append(roomStack, addlExits)
			}
		}
	}

	return nil
}

// Returns true if it appears that the graph should be rebuilt
func (r *RoomGraph) Changed() bool {

	if r.Root == nil {
		return true
	}

	roomNow := LoadRoom(r.Root.RoomId)
	if roomNow == nil {
		return true
	}

	rootRoomId, totalRooms, err := ZoneStats(roomNow.Zone)
	if err != nil {
		return true
	}

	if rootRoomId != r.Root.RoomId {
		return true
	}

	if totalRooms == len(r.trackedRoomIds) {
		slog.Info("RoomGraph::Changed()", "Updated needed, mismatched room counts", "totalRooms", totalRooms, "trackedRoomIds", len(r.trackedRoomIds))
		//	return true
	}

	return false
}

func boundaryCheck(x, y, minX, maxX, minY, maxY int) bool {
	return x >= minX && x < maxX && y >= minY && y < maxY
}
