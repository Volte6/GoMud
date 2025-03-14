package mapper

import (
	"github.com/volte6/gomud/internal/rooms"
)

const defaultMapSymbol = `•`

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

type mapper struct {
	rootRoomId   int              // The room the crawler starts from
	crawlQueue   []int            // A stack of rooms to crawl
	crawledRooms map[int]*mapNode // A look up table of rooms already crawled
}

func NewMapper(rootRoomId int) *mapper {
	return &mapper{
		rootRoomId:   rootRoomId,
		crawledRooms: make(map[int]*mapNode, 100), // pre-allocate 100
	}
}

func (r *mapper) Start() {

	newRoomCrawlCounter := 0

	r.crawlQueue = make([]int, 0, 100) // pre-allocate 100 capacity

	r.crawlQueue = append(r.crawlQueue, r.rootRoomId)
	for len(r.crawlQueue) > 0 {

		roomId := r.crawlQueue[0]

		r.crawlQueue = r.crawlQueue[1:]

		if _, ok := r.crawledRooms[roomId]; ok {
			continue
		}

		node := r.getMapNode(roomId)

		newRoomCrawlCounter++

		// Add to crawled list so we don't revisit it
		r.crawledRooms[node.RoomId] = node

		// Now process it
		for _, exitInfo := range node.Exits {
			if _, ok := r.crawledRooms[exitInfo.RoomId]; ok {
				continue
			}
			r.crawlQueue = append(r.crawlQueue, exitInfo.RoomId)
		}

	}

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

	if len(room.MapSymbol) > 0 {
		mNode.Symbol = []rune(room.MapSymbol)[0]
		if len(room.MapLegend) > 0 {
			mNode.Legend = room.MapLegend
		}
	} else {
		b := room.GetBiome()
		if b.Symbol() != 0 {
			mNode.Symbol = b.Symbol()
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

func (r *mapper) GetMap(centerRoomId int, mapWidth int, mapHeight int, zoomOut bool) mapRender {

	out := newMapRender(mapWidth, mapHeight)

	multiplier := 2
	if zoomOut {
		multiplier = 1
	}

	centerY := mapHeight >> 1
	centerX := mapWidth >> 1
	centerZ := 0

	//	mapRender[centerY][centerX] = 'X'

	startNode := r.crawledRooms[centerRoomId]
	if startNode == nil {
		return out
	}

	visited := make(map[int]map[int]struct{}, 100)
	mappingQueue := make([]WalkStep, 0, 100) // pre-allocate 100 capacity

	mappingQueue = append(mappingQueue, WalkStep{RoomId: centerRoomId, FromPos: positionDelta{centerX, centerY, centerZ, '¿'}, RelativePos: positionDelta{0, 0, 0, '?'}})
	for len(mappingQueue) > 0 {

		roomStep := mappingQueue[0]

		mappingQueue = mappingQueue[1:]

		if _, ok := visited[roomStep.FromRoomId][roomStep.RoomId]; ok {
			continue
		}

		node := r.crawledRooms[roomStep.RoomId]
		if node == nil {
			continue
		}

		adjustedY := roomStep.RelativePos.y*multiplier + centerY
		adjustedX := roomStep.RelativePos.x*multiplier + centerX
		adjustedZ := roomStep.RelativePos.z*multiplier + centerZ

		if adjustedZ == 0 {

			travelDistance := 0
			if roomStep.FromPos.x != adjustedX || roomStep.FromPos.y != adjustedY {

				yDir := 0
				if adjustedY < roomStep.FromPos.y {
					yDir = -1
				} else if adjustedY > roomStep.FromPos.y {
					yDir = 1
				}

				xDir := 0
				if adjustedX < roomStep.FromPos.x {
					xDir = -1
				} else if adjustedX > roomStep.FromPos.x {
					xDir = 1
				}

				y := roomStep.FromPos.y
				x := roomStep.FromPos.x
				for y != adjustedY || x != adjustedX {

					if x != adjustedX {
						x += xDir
					}
					if y != adjustedY {
						y += yDir
					}

					// Don't let it run out of control
					travelDistance++
					if travelDistance > 5 {
						break
					}

					if y == adjustedY && x == adjustedX {
						continue
					}

					if y < 0 || y >= mapHeight || x < 0 || x >= mapWidth {
						continue
					}

					if roomStep.SecretPath {
						out.Render[y][x] = '?'
					} else if roomStep.LockedPath {
						out.Render[y][x] = '⚷'
					} else {
						out.Render[y][x] = roomStep.RelativePos.arrow
					}
				}

			}

			if adjustedY < 0 || adjustedY >= mapHeight {
				continue
			}

			if adjustedX < 0 || adjustedX >= mapWidth {
				continue
			}

			out.Render[adjustedY][adjustedX] = node.Symbol

			if node.Symbol != 0 {
				if _, ok := out.legend[node.Symbol]; !ok {
					out.legend[node.Symbol] = node.Legend
				}
			}

		}

		// Add to crawled list so we don't revisit it
		if _, ok := visited[roomStep.FromRoomId]; !ok {
			visited[roomStep.FromRoomId] = map[int]struct{}{}
			visited[roomStep.FromRoomId][roomStep.RoomId] = struct{}{}

			if _, ok := visited[roomStep.FromRoomId]; !ok {
				visited[roomStep.RoomId] = map[int]struct{}{}
				visited[roomStep.RoomId][roomStep.FromRoomId] = struct{}{}
			}
		}

		// Now process it
		for _, exitInfo := range node.Exits {

			if _, ok := visited[exitInfo.RoomId]; ok {
				continue
			}

			newPos := positionDelta{
				roomStep.RelativePos.x + exitInfo.Direction.x,
				roomStep.RelativePos.y + exitInfo.Direction.y,
				roomStep.RelativePos.z + exitInfo.Direction.z,
				exitInfo.Direction.arrow,
			}
			mappingQueue = append(
				mappingQueue,
				WalkStep{
					RoomId:      exitInfo.RoomId,
					FromRoomId:  node.RoomId,
					SecretPath:  exitInfo.Secret,
					LockedPath:  exitInfo.Locked,
					FromPos:     positionDelta{adjustedX, adjustedY, adjustedZ, 0},
					RelativePos: newPos, //roomStep.RelativePos.Combine(exitInfo.Direction),
				},
			)

		}

	}

	out.Render[centerY][centerX] = 'X'

	return out

}

type WalkStep struct {
	RoomId      int
	FromRoomId  int
	SecretPath  bool
	LockedPath  bool
	FromPos     positionDelta
	RelativePos positionDelta
}
