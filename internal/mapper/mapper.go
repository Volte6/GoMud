package mapper

import (
	"math"

	"github.com/volte6/gomud/internal/rooms"
)

const defaultMapSymbol = '•'

var (
	posDeltas = map[string]positionDelta{
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

type positionDelta struct {
	x     int
	y     int
	z     int
	arrow rune
}

func (p positionDelta) Combine(p2 positionDelta) positionDelta {
	p.x += p2.x
	p.y += p2.y
	p.z += p2.z
	p.arrow = p2.arrow
	return p
}

type RoomGrid struct {
	size       positionDelta
	roomOffset positionDelta
	rooms      [][][]*mapNode // All rooms in a relative position
}

func (r *RoomGrid) initialize(minX, maxX, minY, maxY, minZ, maxZ int) {

	r.size.x = maxX - minX + 1
	r.size.y = maxY - minY + 1
	r.size.z = maxZ - minZ + 1

	r.roomOffset.x = minX * -1
	r.roomOffset.y = minY * -1
	r.roomOffset.z = minZ * -1

	r.rooms = make([][][]*mapNode, r.size.z)
	for z := 0; z < r.size.z; z++ {
		r.rooms[z] = make([][]*mapNode, r.size.y)
		for y := 0; y < r.size.y; y++ {
			r.rooms[z][y] = make([]*mapNode, r.size.x)
		}
	}

}

func (r *RoomGrid) addNode(n *mapNode) {

	adjX := n.Pos.x + r.roomOffset.x
	adjY := n.Pos.y + r.roomOffset.y
	adjZ := n.Pos.z + r.roomOffset.z

	r.rooms[adjZ][adjY][adjX] = n

}

type mapper struct {
	rootRoomId   int              // The room the crawler starts from
	crawlQueue   []crawlRoom      // A stack of rooms to crawl
	crawledRooms map[int]*mapNode // A look up table of rooms already crawled
	gridOffsets  positionDelta

	roomGrid RoomGrid
}

func NewMapper(rootRoomId int) *mapper {
	return &mapper{
		rootRoomId:   rootRoomId,
		crawledRooms: make(map[int]*mapNode, 100), // pre-allocate 100
		roomGrid: RoomGrid{
			rooms: [][][]*mapNode{},
		},
	}
}

type crawlRoom struct {
	RoomId int
	Pos    positionDelta // Its x/y/z position relative to the root node
}

func (r *mapper) Start() {

	minX, maxX, minY, maxY, minZ, maxZ := 0, 0, 0, 0, 0, 0

	r.crawlQueue = make([]crawlRoom, 0, 100) // pre-allocate 100 capacity

	//var lastNode *mapNode = nil
	r.crawlQueue = append(r.crawlQueue, crawlRoom{RoomId: r.rootRoomId, Pos: positionDelta{}})
	for len(r.crawlQueue) > 0 {

		roomNow := r.crawlQueue[0]

		r.crawlQueue = r.crawlQueue[1:]

		if _, ok := r.crawledRooms[roomNow.RoomId]; ok {
			continue
		}

		node := r.getMapNode(roomNow.RoomId)
		node.Pos = roomNow.Pos

		// Add to crawled list so we don't revisit it
		r.crawledRooms[node.RoomId] = node

		// Now process it
		for _, exitInfo := range node.Exits {
			if _, ok := r.crawledRooms[exitInfo.RoomId]; ok {
				continue
			}

			newCrawl := crawlRoom{
				RoomId: exitInfo.RoomId,
				Pos:    roomNow.Pos.Combine(exitInfo.Direction),
			}

			if newCrawl.Pos.x < minX {
				minX = newCrawl.Pos.x
			} else if newCrawl.Pos.x > maxX {
				maxX = newCrawl.Pos.x
			}

			if newCrawl.Pos.y < minY {
				minY = newCrawl.Pos.y
			} else if newCrawl.Pos.y > maxY {
				maxY = newCrawl.Pos.y
			}

			if newCrawl.Pos.y < minZ {
				minZ = newCrawl.Pos.z
			} else if newCrawl.Pos.y > maxZ {
				maxZ = newCrawl.Pos.z
			}

			r.crawlQueue = append(r.crawlQueue, newCrawl)
		}

	}

	// calculate the final array length.

	r.roomGrid.initialize(minX, maxX, minY, maxY, minZ, maxZ)

	for _, node := range r.crawledRooms {
		r.roomGrid.addNode(node)
	}
}

func (r *mapper) GetMap(centerRoomId int, mapWidth int, mapHeight int, zoomLevel int) mapRender {

	if zoomLevel < 0 {
		zoomLevel = 0
	}
	zoomLevel++

	/*
		+-----------+
		|     ••&   |
		|    %••••••|
		|⌂⌂⌂⌂%••⌂⌂⌂⌂|
		|•••••••••••|
		|  •★•• $⌂$ |
		|•••••@•••••|
		|  •  • $ I•|
		|•••  ••••• |
		|••P• ••+ • |
		| ••  ••••• |
		|     ••••••|
		+-----------+

		+-----------+
		|• • • • • •|
		|           |
		|★ • •   $ ⌂|
		|           |
		|• • @ • • •|
		|           |
		|    •   $  |
		|           |
		|    • • • •|
		|           |
		|•   • • +  |
		+-----------+
	*/
	out := newMapRender(mapWidth, mapHeight)

	node := r.crawledRooms[centerRoomId]

	if node == nil {
		return out
	}

	srcPos := positionDelta{
		x: node.Pos.x + r.roomGrid.roomOffset.x - int(math.Ceil(float64(mapWidth)/float64(zoomLevel)))>>1,
		y: node.Pos.y + r.roomGrid.roomOffset.y - int(math.Ceil(float64(mapHeight)/float64(zoomLevel)))>>1,
		z: 0,
	}

	dstPos := positionDelta{
		x: 0,
		y: 0,
		z: 0,
	}

	srcEndX := srcPos.x + int(math.Ceil(float64(mapWidth)/float64(zoomLevel)))
	srcEndY := srcPos.y + int(math.Ceil(float64(mapHeight)/float64(zoomLevel)))

	// Make sure we don't try and draw from beyond the grid
	if srcPos.x < 0 {
		dstPos.x = (srcPos.x * -1) * zoomLevel
		srcPos.x = 0
	}

	if srcPos.y < 0 {
		dstPos.y = srcPos.y * -1 * zoomLevel
		srcPos.y = 0
	}

	if srcEndX > r.roomGrid.size.x-1 {
		srcEndX = r.roomGrid.size.x - 1

	}

	if srcEndY > r.roomGrid.size.y-1 {
		srcEndY = r.roomGrid.size.y - 1
	}

	drawX, drawY := 0, 0
	z := 0

	for y := srcPos.y; y < srcEndY; y++ {

		drawX = 0
		for x := srcPos.x; x < srcEndX; x++ {

			if node = r.roomGrid.rooms[z][y][x]; node != nil {

				if _, ok := out.legend[node.Symbol]; !ok {
					out.legend[node.Symbol] = node.Legend
				}

				if centerRoomId == node.RoomId {
					out.Render[dstPos.y+drawY][dstPos.x+drawX] = '@'
				} else {
					out.Render[dstPos.y+drawY][dstPos.x+drawX] = node.Symbol
				}

				// draw any exits... only if zoom level is > 1
				if zoomLevel > 1 {

					xStart, yStart := dstPos.x+drawX, dstPos.y+drawY
					for _, exitInfo := range node.Exits {

						maxSteps := zoomLevel

						xStepDir := 0
						if exitInfo.Direction.x < 0 {
							if exitInfo.Direction.x < -1 {
								maxSteps += 1
							}
							xStepDir = -1
						} else if exitInfo.Direction.x > 0 {
							if exitInfo.Direction.x > 1 {
								maxSteps += 1
							}
							xStepDir = 1
						}

						yStepDir := 0
						if exitInfo.Direction.y < 0 {
							if exitInfo.Direction.y < -1 {
								maxSteps += 1
							}
							yStepDir = -1
						} else if exitInfo.Direction.y > 0 {
							if exitInfo.Direction.y > 1 {
								maxSteps += 1
							}
							yStepDir = 1
						}

						drawX2, drawY2 := 0, 0
						for step := 1; step < maxSteps; step++ {

							drawY2 = yStart + yStepDir*step
							drawX2 = xStart + xStepDir*step

							if drawY2 < 0 || drawY2 >= mapHeight {
								continue
							}
							if drawX2 < 0 || drawX2 >= mapWidth {
								continue
							}

							if exitInfo.Secret {
								out.Render[drawY2][drawX2] = '?'
							} else if exitInfo.Locked {
								out.Render[drawY2][drawX2] = '⚷'
							} else {
								if out.Render[drawY2][drawX2] == ' ' {
									out.Render[drawY2][drawX2] = exitInfo.Direction.arrow
								}
							}

						}

					}

				}
			}
			drawX += zoomLevel
		}
		drawY += zoomLevel
	}

	return out

}

func (r *mapper) getMapNode(roomId int) *mapNode {

	if r, ok := r.crawledRooms[roomId]; ok {
		return r
	}

	room := rooms.LoadRoom(roomId)
	if room == nil {
		return nil
	}

	mNode := &mapNode{
		RoomId:      room.RoomId,
		Exits:       make(map[string]nodeExit, 2), // assume there will be on average 2 exits per room
		SecretExits: make(map[string]struct{}),
	}

	if room.MapSymbol != `` {
		mNode.Symbol = []rune(room.MapSymbol)[0]
		if room.MapLegend != `` {
			mNode.Legend = room.MapLegend
		}
	} else {
		b := room.GetBiome()
		if b.Symbol() != 0 {
			mNode.Symbol = b.Symbol()
		} else {
			mNode.Symbol = defaultMapSymbol
		}
		if b.Name() != `` {
			mNode.Legend = b.Name()
		}
	}

	for exitName, exitInfo := range room.Exits {
		exitNode := nodeExit{
			RoomId: exitInfo.RoomId,
			Secret: exitInfo.Secret,
			Locked: exitInfo.HasLock(),
		}

		if d, ok := posDeltas[exitInfo.MapDirection]; ok {
			exitNode.Direction = d
		} else if d, ok := posDeltas[exitName]; ok {
			exitNode.Direction = d
		} else {
			continue
		}

		mNode.Exits[exitName] = exitNode
	}

	return mNode
}

type WalkStep struct {
	RoomId      int
	FromRoomId  int
	SecretPath  bool
	LockedPath  bool
	FromPos     positionDelta
	RelativePos positionDelta
}
