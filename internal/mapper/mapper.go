package mapper

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/users"
)

const (
	defaultMapSymbol = '•'
	SecretSymbol     = '?'
	LockedSymbol     = '⚷'
)

var (
	ErrOutOfBounds  = errors.New(`out of bounds`)
	ErrRoomNotFound = errors.New(`room not found`)

	compassDirections = map[string]struct{}{
		"north":     struct{}{},
		"south":     struct{}{},
		"west":      struct{}{},
		"east":      struct{}{},
		"northwest": struct{}{},
		"northeast": struct{}{},
		"southwest": struct{}{},
		"southeast": struct{}{},
		"down":      struct{}{},
		"up":        struct{}{},
	}

	// TODO: Refactor this and remove string identifier from room data.
	posDeltas = map[string]positionDelta{
		"north":     {0, -1, 0, '│'},
		"south":     {0, 1, 0, '│'},
		"west":      {-1, 0, 0, '─'},
		"east":      {1, 0, 0, '─'},
		"northwest": {-1, -1, 0, '╲'},
		"northeast": {1, -1, 0, '╱'},
		"southwest": {-1, 1, 0, '╱'},
		"southeast": {1, 1, 0, '╲'},
		"down":      {0, 0, -1, 'v'},
		"up":        {0, 0, 1, '^'},

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

	mapperZoneCache = map[string]*mapper{}
)

func GetDelta(exitName string) (x, y, z int) {
	if delta, ok := posDeltas[exitName]; ok {
		return delta.x, delta.y, delta.z
	}
	return 0, 0, 0
}

func IsCompassDirection(exitName string) bool {
	_, ok := compassDirections[exitName]
	return ok
}

func GetDirectionDeltaNames() []string {
	ret := []string{}
	for name, _ := range posDeltas {
		ret = append(ret, name)
	}
	return ret
}

func IsValidDirection(directionName string) bool {
	_, ok := posDeltas[directionName]
	return ok
}

func GetZoneMapper(zoneName string, forceRefresh ...bool) *mapper {

	doRefresh := len(forceRefresh) > 0 && forceRefresh[0]

	if !doRefresh {
		if _, ok := mapperZoneCache[zoneName]; ok {
			return mapperZoneCache[zoneName]
		}
	}

	roomId, err := rooms.GetZoneRoot(zoneName)
	if err != nil {
		return nil
	}

	for zName, mapper := range mapperZoneCache {
		if mapper.HasRoom(roomId) {

			if doRefresh {
				delete(mapperZoneCache, zName)
				zoneName = zName
			} else {
				return mapper
			}
		}
	}

	// not found. Will need to create one.
	tStart := time.Now()

	m := NewMapper(roomId)
	m.Start()

	mudlog.Info("New Mapper", "zone", zoneName, "time taken", time.Since(tStart))

	mapperZoneCache[zoneName] = m

	return m
}

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

	// Don't redo.
	if len(r.crawledRooms) > 0 {
		return
	}

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

			if newCrawl.Pos.z < minZ {
				minZ = newCrawl.Pos.z
			} else if newCrawl.Pos.z > maxZ {
				maxZ = newCrawl.Pos.z
			}

			r.crawlQueue = append(r.crawlQueue, newCrawl)
		}

	}

	r.crawlQueue = nil

	// calculate the final array length.

	r.roomGrid.initialize(minX, maxX, minY, maxY, minZ, maxZ)

	for _, node := range r.crawledRooms {
		r.roomGrid.addNode(node)
	}
}

func (r *mapper) HasRoom(roomId int) bool {
	return r.crawledRooms[roomId] != nil
}

// Get the roomId at a given coordinate
func (r *mapper) GetRoomId(x, y, z int) (roomId int, err error) {

	z += r.roomGrid.roomOffset.z
	if z < 0 {
		return 0, ErrOutOfBounds
	}
	if z >= len(r.roomGrid.rooms) {
		return 0, ErrOutOfBounds
	}

	y += r.roomGrid.roomOffset.y
	if y < 0 {
		return 0, ErrOutOfBounds
	}
	if y >= len(r.roomGrid.rooms[z]) {
		return 0, ErrOutOfBounds
	}

	x += r.roomGrid.roomOffset.x
	if x < 0 {
		return 0, ErrOutOfBounds
	}
	if x >= len(r.roomGrid.rooms[z][y]) {
		return 0, ErrOutOfBounds
	}

	if checkNode := r.roomGrid.rooms[z][y][x]; checkNode != nil {
		return checkNode.RoomId, nil
	}

	return 0, ErrRoomNotFound
}

// Get the coordinates of a given roomId
func (r *mapper) GetCoordinates(roomId int) (x, y, z int, err error) {

	node, ok := r.crawledRooms[roomId]
	if !ok {
		return 0, 0, 0, ErrRoomNotFound
	}

	return node.Pos.x, node.Pos.y, node.Pos.z, nil
}

// Finds the first room in a given direction
// Allowed directions:
func (r *mapper) FindAdjacentRoom(centerRoomId int, direction string, limitDistance ...int) (roomId int, distance int) {

	startNode := r.crawledRooms[centerRoomId]
	if startNode == nil {
		return 0, 0
	}

	// Make sure there isn't just an exit there already.
	if exitNode, ok := startNode.Exits[direction]; ok {

		if exitNode.Direction.x != 0 {
			return exitNode.RoomId, int(math.Abs(float64(exitNode.Direction.x)))
		} else if exitNode.Direction.y != 0 {
			return exitNode.RoomId, int(math.Abs(float64(exitNode.Direction.y)))
		} else if exitNode.Direction.z != 0 {
			return exitNode.RoomId, int(math.Abs(float64(exitNode.Direction.z)))
		}

		return exitNode.RoomId, 0
	}

	dirParts := strings.FieldsFunc(direction, func(r rune) bool {
		return !unicode.IsLetter(r)
	})

	if len(dirParts) > 0 {
		direction = dirParts[0]
	}

	direction = strings.ToLower(direction)

	checkDirection, ok := posDeltas[direction]
	if !ok {
		return 0, 0
	}

	x, y, z := startNode.Pos.x, startNode.Pos.y, startNode.Pos.z
	steps := 0
	for {

		x += checkDirection.x
		y += checkDirection.y
		z += checkDirection.z

		if len(limitDistance) > 0 {
			steps++
			if steps > limitDistance[0] {
				return 0, 0
			}
		}

		roomId, err := r.GetRoomId(x, y, z)
		if err != nil {

			if err == ErrOutOfBounds {
				return 0, 0
			}

			// Otherwise keep searching until out of bounds or no error.
			continue
		}

		// Get the number of rooms away. Since we are only taking single steps, can use any non zero value.
		distance := 0
		if x != 0 {
			distance = int(math.Abs(float64(x)))
		} else if y != 0 {
			distance = int(math.Abs(float64(y)))
		} else if z != 0 {
			distance = int(math.Abs(float64(z)))
		}

		return roomId, distance
	}

}

// BFS Crawls from the center position as though the players POV
// So it won't follow locked rooms, secrets, etc.
func (r *mapper) GetLimitedMap(centerRoomId int, c Config) mapRender {

	if c.ZoomLevel < 0 {
		c.ZoomLevel = 0
	}
	c.ZoomLevel++

	out := newMapRender(c.Width, c.Height)

	centerX, centerY := c.Width>>1, c.Height>>1

	// Start drawing from the center of the output canvas
	dstPos := positionDelta{
		x: centerX,
		y: centerY,
		z: 0,
	}

	node := r.crawledRooms[centerRoomId]
	if node == nil {
		return out
	}

	r.crawlQueue = make([]crawlRoom, 0, 20) // pre-allocate 20 capacity

	drawX, drawY := 0, 0

	// Special additional crawl tracker.
	// This is because we're doing a unique crawl right now, independent of the "global crawl"
	tmpCrawledTracker := map[int]struct{}{}

	//var lastNode *mapNode = nil
	r.crawlQueue = append(r.crawlQueue, crawlRoom{RoomId: node.RoomId, Pos: dstPos})

	var symbol rune
	var legend string

	for len(r.crawlQueue) > 0 {

		roomNow := r.crawlQueue[0]
		r.crawlQueue = r.crawlQueue[1:]

		if _, ok := tmpCrawledTracker[roomNow.RoomId]; ok {
			continue
		}

		dstPos = roomNow.Pos

		node := r.getMapNode(roomNow.RoomId)

		symbol = node.Symbol
		legend = node.Legend

		if c.symbolOverrides != nil {
			if override, ok := c.symbolOverrides[node.RoomId]; ok {
				symbol = override.Symbol
				legend = override.Legend
			}
		}

		if dstPos.z == 0 && dstPos.x >= 0 && dstPos.y >= 0 && dstPos.x < c.Width && dstPos.y < c.Height {
			// Draw the room to the output
			out.Render[dstPos.y][dstPos.x] = symbol
			if _, ok := out.legend[symbol]; !ok {
				out.legend[symbol] = legend
			}
		}

		// Add to crawled list so we don't revisit it
		tmpCrawledTracker[node.RoomId] = struct{}{}

		// Now process it
		skip := false
		for _, exitInfo := range node.Exits {
			if _, ok := tmpCrawledTracker[exitInfo.RoomId]; ok {
				continue
			}

			skip = false

			// Draw exit (if applicable) and add it to the crawlQueue
			// Do not crawl if the critia is not met for the exit.
			// For example: a secret exit that the user has not visited.

			showLocked := false
			if exitInfo.LockDifficulty > 0 {

				if c.UserId > 0 {

					user := users.GetByUserId(c.UserId)
					if user != nil {
						hasKey, hasSequence := user.Character.HasKey(exitInfo.LockId, exitInfo.LockDifficulty)
						showLocked = hasKey || hasSequence
					}
				} else if c.UserId < 0 { // Debug
					showLocked = true
				}
			}

			if exitInfo.Secret {
				targetRoom := rooms.LoadRoom(exitInfo.RoomId)
				if targetRoom == nil {
					continue
				}

				if c.UserId >= 0 {
					if !targetRoom.HasVisited(c.UserId, rooms.VisitorUser) {
						continue
					}
				}
			}

			// Dont' draw if z-plane has moved
			if dstPos.z == 0 && exitInfo.Direction.z == 0 {

				maxSteps := c.ZoomLevel

				xStepDir := 0
				if exitInfo.Direction.x < 0 {
					if exitInfo.Direction.x < -1 {
						maxSteps = c.ZoomLevel * -exitInfo.Direction.x
					}
					xStepDir = -1
				} else if exitInfo.Direction.x > 0 {
					if exitInfo.Direction.x > 1 {
						maxSteps = c.ZoomLevel * exitInfo.Direction.x
					}
					xStepDir = 1
				}

				yStepDir := 0
				if exitInfo.Direction.y < 0 {
					if exitInfo.Direction.y < -1 {
						maxSteps = c.ZoomLevel * -exitInfo.Direction.y
					}
					yStepDir = -1
				} else if exitInfo.Direction.y > 0 {
					if exitInfo.Direction.y > 1 {
						maxSteps = c.ZoomLevel * exitInfo.Direction.y
					}
					yStepDir = 1
				}

				for step := 1; step < maxSteps; step++ {

					drawX = dstPos.x + xStepDir*step
					drawY = dstPos.y + yStepDir*step

					if drawX >= 0 && drawY >= 0 && drawX < c.Width && drawY < c.Height {

						if exitInfo.Secret {
							out.Render[drawY][drawX] = SecretSymbol
							if _, ok := out.legend[SecretSymbol]; !ok {
								out.legend[SecretSymbol] = `Secret`
							}
						} else if exitInfo.LockDifficulty > 0 {
							out.Render[drawY][drawX] = LockedSymbol
							if _, ok := out.legend[LockedSymbol]; !ok {
								out.legend[LockedSymbol] = `Locked`
							}
							if !showLocked {
								skip = true
								break
							}
						} else {
							out.Render[drawY][drawX] = exitInfo.Direction.arrow
						}

					}
				}

			}

			if !skip {
				newCrawl := crawlRoom{
					RoomId: exitInfo.RoomId,
					Pos:    dstPos,
				}

				for i := 0; i < c.ZoomLevel; i++ {
					newCrawl.Pos = newCrawl.Pos.Combine(exitInfo.Direction)
				}

				r.crawlQueue = append(r.crawlQueue, newCrawl)
			}

		}

	}

	return out
}

// Returns a fully rendered map of the area
func (r *mapper) GetFullMap(centerRoomId int, c Config) mapRender {

	if c.ZoomLevel < 0 {
		c.ZoomLevel = 0
	}
	c.ZoomLevel++

	out := newMapRender(c.Width, c.Height)

	node := r.crawledRooms[centerRoomId]

	if node == nil {
		return out
	}

	srcPos := positionDelta{
		x: node.Pos.x + r.roomGrid.roomOffset.x - int(math.Ceil(float64(c.Width)/float64(c.ZoomLevel)))>>1,
		y: node.Pos.y + r.roomGrid.roomOffset.y - int(math.Ceil(float64(c.Height)/float64(c.ZoomLevel)))>>1,
		z: node.Pos.z + r.roomGrid.roomOffset.z,
	}

	dstPos := positionDelta{
		x: 0,
		y: 0,
		z: 0,
	}

	srcEndX := srcPos.x + int(math.Ceil(float64(c.Width)/float64(c.ZoomLevel)))
	srcEndY := srcPos.y + int(math.Ceil(float64(c.Height)/float64(c.ZoomLevel)))

	// Make sure we don't try and draw from beyond the grid
	if srcPos.x < 0 {
		dstPos.x = (srcPos.x * -1) * c.ZoomLevel
		srcPos.x = 0
	}

	if srcPos.y < 0 {
		dstPos.y = srcPos.y * -1 * c.ZoomLevel
		srcPos.y = 0
	}

	if srcEndX > r.roomGrid.size.x-1 {
		srcEndX = r.roomGrid.size.x - 1
	}

	if srcEndY > r.roomGrid.size.y-1 {
		srcEndY = r.roomGrid.size.y - 1
	}

	drawX, drawY := 0, 0
	z := srcPos.z

	var symbol rune
	var legend string

	for y := srcPos.y; y < srcEndY; y++ {

		drawX = 0
		for x := srcPos.x; x < srcEndX; x++ {

			if node = r.roomGrid.rooms[z][y][x]; node != nil {

				symbol = node.Symbol
				legend = node.Legend

				if c.symbolOverrides != nil {
					if override, ok := c.symbolOverrides[node.RoomId]; ok {
						symbol = override.Symbol
						legend = override.Legend
					}
				}

				out.Render[dstPos.y+drawY][dstPos.x+drawX] = symbol

				if _, ok := out.legend[symbol]; !ok {
					out.legend[symbol] = legend
				}

				// draw exits
				xStart, yStart := dstPos.x+drawX, dstPos.y+drawY
				for _, exitInfo := range node.Exits {

					maxSteps := c.ZoomLevel

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

						if drawY2 < 0 || drawY2 >= c.Height {
							continue
						}
						if drawX2 < 0 || drawX2 >= c.Width {
							continue
						}

						if exitInfo.Secret {
							out.Render[drawY2][drawX2] = SecretSymbol
							if _, ok := out.legend[SecretSymbol]; !ok {
								out.legend[SecretSymbol] = `Secret`
							}
						} else if exitInfo.LockDifficulty > 0 {
							out.Render[drawY2][drawX2] = LockedSymbol
							if _, ok := out.legend[LockedSymbol]; !ok {
								out.legend[LockedSymbol] = `Locked`
							}
						} else {
							out.Render[drawY2][drawX2] = exitInfo.Direction.arrow
						}

					}

				}

			}
			drawX += c.ZoomLevel
		}
		drawY += c.ZoomLevel
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
			RoomId:         exitInfo.RoomId,
			Secret:         exitInfo.Secret,
			LockDifficulty: int(exitInfo.Lock.Difficulty),
		}

		if exitNode.LockDifficulty > 0 {
			exitNode.LockId = fmt.Sprintf(`%d-%s`, room.RoomId, exitName)
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

func PreCacheMaps() {

	// Sort the rooms by roomId before precaching.
	// This ensures a somewhat predictable inferred coordinate system across
	// MUD server starts.
	type ZoneDetails struct {
		Name       string
		RootRoomId int
	}

	allZones := []ZoneDetails{}

	for _, name := range rooms.GetAllZoneNames() {
		rootRoomId, _ := rooms.GetZoneRoot(name)
		allZones = append(allZones, ZoneDetails{
			Name:       name,
			RootRoomId: rootRoomId,
		})
	}

	sort.Slice(allZones, func(i, j int) bool {
		return allZones[i].RootRoomId < allZones[j].RootRoomId
	})

	for _, zInfo := range allZones {
		GetZoneMapper(zInfo.Name)
	}
}
