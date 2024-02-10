package rooms

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/users"
)

var (

	// Overrides legend for stuff that appears on the map
	legendOverrides = map[string]string{
		`$`: `Shop`,
		`★`: `Bank`,
		`G`: `Gate`,
		`✗`: `Target`,
		`%`: `Trainer`,
	}
)

type MapBorder struct {
	Top    string
	Mid    []string
	Bottom string
}

type MapData struct {
	Title        string
	Zone         string
	Width        int
	Height       int
	DisplayLines []string
	Legend       map[string]string
	LegendWidth  int
	LeftBorder   MapBorder
	MidBorder    MapBorder
	RightBorder  MapBorder
}

func GenerateZoneMap(zone string, roomId int, userId int, limitXDistance int, limitYDistance int, mapMode MapMode) *RoomGraph {

	if roomId == 0 {
		roomId, _ = GetZoneRoot(zone)
	}

	povUserId := userId
	user := users.GetByUserId(userId)

	// If a user with a level greater than possible or a non-user centric map, then don't restrict by pov
	if user == nil || user.Character.GetSkillLevel(skills.Map) > 4 {
		if mapMode != MapModeTracking {
			povUserId = 0
		}
	}

	var overrideSymbols map[int]rune = nil

	if mapMode == MapModeTracking {
		overrideSymbols = make(map[int]rune)
		overrideSymbols[user.Character.RoomId] = '✗'
	}

	xDistance := 500
	yDistance := 500

	if limitXDistance != 0 {
		xDistance = limitXDistance
	}
	if limitYDistance != 0 {
		yDistance = limitYDistance
	}

	rGraph := NewRoomGraph(xDistance, yDistance, povUserId, mapMode)
	err := rGraph.Build(roomId, overrideSymbols)
	if err != nil {
		panic(err)
	}

	return rGraph
}

func DrawZoneMap(rGraph *RoomGraph, title string, mapDisplayWidth int, mapDisplayHeight int) (MapData, error) {

	map2D /*cRoomX, cRoomY*/, _, _ := rGraph.Generate2DMap(mapDisplayWidth/2, mapDisplayHeight/2, rGraph.Root.RoomId)

	/*
		fmt.Println()
		fmt.Println("Generated a 2D Map To String")
		fmt.Printf("%s\n", map2D.String())
		fmt.Println()
	*/

	yPaddingTop := mapDisplayHeight - (len(map2D) * 2)
	if yPaddingTop < 0 {
		yPaddingTop = 0
	}
	yPaddingBottom := int(math.Floor(float64(yPaddingTop) / 2))
	yPaddingTop -= yPaddingBottom

	xPaddingLeft := mapDisplayWidth - (len(map2D[0]) * 2)
	if xPaddingLeft < 0 {
		xPaddingLeft = 0
	}
	xPaddingRight := int(math.Floor(float64(xPaddingLeft) / 2))
	xPaddingLeft -= xPaddingRight

	mapLines := make([]string, 0, mapDisplayHeight)

	// Expand it to the full size

	// First build the final map
	finalMap := make([][]rune, mapDisplayHeight)
	for y := 0; y < mapDisplayHeight; y++ {
		finalMap[y] = make([]rune, mapDisplayWidth)
	}

	//fmt.Println("mapDisplayWidth", mapDisplayWidth, "rGraph.width", rGraph.width, "divX", divX)

	// Min and Max render is half the map heieght/width because of rooms connections
	miniLenX, miniLenY := mapDisplayWidth>>1, mapDisplayHeight>>1

	legend := make(map[string]string)
	var node *Map2DNode = nil

	// Loop through the request map (half size)
	// Fill in details, every other line should have room connection info
	for y := 0; y < mapDisplayHeight; y++ {

		for x := 0; x < mapDisplayWidth; x++ {

			miniX := x >> 1
			miniY := y >> 1

			// If it's an even line and event col, fill in the room details from the mini map
			if y%2 == 0 && miniY < miniLenY {

				if x%2 == 0 && miniX < miniLenX {

					node = map2D[miniY][miniX]

					// Nothing found here, put ' '
					if node != nil {

						// check all exits to draw an arrow in that direction
						for _, dirDelta := range node.Exits {
							if y+dirDelta.Dy >= 0 && y+dirDelta.Dy < mapDisplayHeight && x+dirDelta.Dx >= 0 && x+dirDelta.Dx < mapDisplayWidth {

								//
								// This section here prunes the room connectors that connect to empty space
								// due to reaching the ends of a map range or something
								//
								if miniY+dirDelta.Dy >= 0 && miniY+dirDelta.Dy < len(map2D) {
									if miniX+dirDelta.Dx >= 0 && miniX+dirDelta.Dx < len(map2D[0]) {
										if map2D[miniY+dirDelta.Dy][miniX+dirDelta.Dx] == nil {
											continue
										}
									}
								}

								iterations := dirDelta.Dx
								if iterations == 0 {
									iterations = dirDelta.Dy
								}

								// invert negative iterations
								if iterations < 0 {
									iterations *= -1
								}

								xTravel := 0
								yTravel := 0
								if dirDelta.Dx > 0 {
									xTravel = 1
								} else if dirDelta.Dx < 0 {
									xTravel = -1
								}
								if dirDelta.Dy > 0 {
									yTravel = 1
								} else if dirDelta.Dy < 0 {
									yTravel = -1
								}

								iterations = iterations*2 - 1
								for i := 1; i <= iterations; i++ {
									newY := y + yTravel*i
									newX := x + xTravel*i
									if newY >= 0 && newY < len(finalMap) {
										if newX >= 0 && newX < len(finalMap[0]) {
											if finalMap[newY][newX] == 0 {
												finalMap[newY][newX] = rune(dirDelta.Arrow)
											}
										}
									}
								}

								/*
																	if finalMap[y+dirDelta.Dy][x+dirDelta.Dx] == 0 {
										finalMap[y+dirDelta.Dy][x+dirDelta.Dx] = rune(dirDelta.Arrow)
									}
								*/

							}
						}

						// check all secret exits to mark as secret.
						for _, dirDelta := range node.SecretExits {
							if y+dirDelta.Dy > 0 && y+dirDelta.Dy < mapDisplayHeight && x+dirDelta.Dx > 0 && x+dirDelta.Dx < mapDisplayWidth {
								finalMap[y+dirDelta.Dy][x+dirDelta.Dx] = '?'
								legend["?"] = "Secret"
							}
						}

						finalMap[y][x] = rune(node.Symbol)
						legend[string(node.Symbol)] = node.Legend

					} else if finalMap[y][x] == 0 {
						// No node found at this spot, empty space for a future node!
						finalMap[y][x] = ' '
					}
				} else if finalMap[y][x] == 0 {
					// This is an in between column
					finalMap[y][x] = ' '
				}

			} else if finalMap[y][x] == 0 {
				finalMap[y][x] = ' '
			}

		}
	}

	// now build the lines
	line := strings.Builder{}
	for y := 0; y < mapDisplayHeight-1; y++ { // -1 so that the final drawn line includes rooms, not just connections
		line.Reset()
		for x := 0; x < mapDisplayWidth-1; x++ { // -1 so that the final drawn line includes rooms, not just connections
			line.WriteRune(finalMap[y][x])
		}
		mapLines = append(mapLines, line.String())
	}

	// ask for half the height of the visible area, since roads take up a line, so will be twice the size.

	mapData := MapData{
		Title: title,
		//Zone:         zone,
		DisplayLines: mapLines,
		Height:       len(mapLines),
		Width:        utf8.RuneCountInString(mapLines[0]),
		Legend:       legend, //make(map[string]string),
		LegendWidth:  utf8.RuneCountInString(mapLines[0]),
		LeftBorder: MapBorder{
			Top:    ".-=~=-.",
			Mid:    []string{"( _ __)", "(__  _)"},
			Bottom: "`-._.-'",
		},
		MidBorder: MapBorder{
			Top:    "-._.-=",
			Bottom: "-._.-=",
		},
		RightBorder: MapBorder{
			Top:    ".-=~=-.",
			Mid:    []string{"( _ __)", "(__  _)"},
			Bottom: "`-._.-'",
		},
	}

	// Override for smaller maps
	if mapData.Width < 15 {
		mapData.LeftBorder = MapBorder{
			Top:    "",
			Mid:    []string{"│", "│"},
			Bottom: "└",
		}
		mapData.MidBorder = MapBorder{
			Top:    `─────`,
			Bottom: `─────`,
		}
		mapData.RightBorder = MapBorder{
			Top:    "",
			Mid:    []string{"│", "│"},
			Bottom: "┘",
		}
	} else if mapData.Width < 20 {
		mapData.LeftBorder = MapBorder{
			Top:    ".-~-.",
			Mid:    []string{"( __)", "(__ )"},
			Bottom: "`-_-'",
		}
		mapData.MidBorder = MapBorder{
			Top:    "-._.-=",
			Bottom: "-._.-=",
		}
		mapData.RightBorder = MapBorder{
			Top:    ".-~-.",
			Mid:    []string{"( __)", "(__ )"},
			Bottom: "`-_-'",
		}
	}

	// Apply any overrides
	for sym, txt := range legendOverrides {
		if _, ok := legend[sym]; ok {
			legend[sym] = txt
		}
	}

	for i := 0; i < mapData.Height; i++ {
		for sym, txt := range legend {
			if strings.Contains(mapData.DisplayLines[i], sym) {
				txtLc := strings.ToLower(txt)
				mapData.DisplayLines[i] = strings.Replace(mapData.DisplayLines[i], sym, fmt.Sprintf(`<ansi fg="map-room"><ansi fg="map-%s" bg="mapbg-%s">%s</ansi></ansi>`, txtLc, txtLc, sym), -1)
			}
		}
	}

	return mapData, nil
}

func DrawZoneMapWide(rGraph *RoomGraph, title string, mapDisplayWidth int, mapDisplayHeight int) (MapData, error) {

	map2D, _, _ /*cRoomX, cRoomY*/ := rGraph.Generate2DMap(mapDisplayWidth, mapDisplayHeight, rGraph.Root.RoomId)

	/*
		fmt.Println()
		fmt.Println("Generated a 2D Map To String")
		fmt.Printf("%s\n", map2D.String())
		fmt.Println()
	*/

	yPaddingTop := mapDisplayHeight - (len(map2D) * 2)
	if yPaddingTop < 0 {
		yPaddingTop = 0
	}
	yPaddingBottom := int(math.Floor(float64(yPaddingTop) / 2))
	yPaddingTop -= yPaddingBottom

	xPaddingLeft := mapDisplayWidth - (len(map2D[0]) * 2)
	if xPaddingLeft < 0 {
		xPaddingLeft = 0
	}
	xPaddingRight := int(math.Floor(float64(xPaddingLeft) / 2))
	xPaddingLeft -= xPaddingRight

	mapLines := make([]string, 0, mapDisplayHeight)

	// Expand it to the full size

	// First build the final map
	finalMap := make([][]rune, mapDisplayHeight)
	for y := 0; y < mapDisplayHeight; y++ {
		finalMap[y] = make([]rune, mapDisplayWidth)
	}

	legend := make(map[string]string)
	var node *Map2DNode = nil
	// Loop through the request map (half size)
	// Fill in details, every other line should have room connection info
	for y := 0; y < mapDisplayHeight; y++ {

		for x := 0; x < mapDisplayWidth; x++ {

			node = map2D[y][x]
			// Nothing found here, put ' '
			if node != nil {
				if node.RoomId == rGraph.Root.RoomId {
					finalMap[y][x] = '@'
				} else {
					finalMap[y][x] = rune(node.Symbol)
				}
				legend[string(node.Symbol)] = node.Legend
			} else if finalMap[y][x] == 0 {
				// No node found at this spot, empty space for a future node!
				finalMap[y][x] = ' '
			}

		}
	}

	// now build the lines
	//mapLines := make([]string, 0, mapDisplayHeight)
	line := strings.Builder{}
	for y := 0; y < mapDisplayHeight; y++ {
		line.Reset()
		for x := 0; x < mapDisplayWidth; x++ {
			line.WriteRune(finalMap[y][x])
		}
		mapLines = append(mapLines, line.String())
	}

	// ask for half the height of the visible area, since roads take up a line, so will be twice the size.

	mapData := MapData{
		Title:        title,
		DisplayLines: mapLines,
		Height:       len(mapLines),
		Width:        utf8.RuneCountInString(mapLines[0]),
		Legend:       legend,
		LeftBorder: MapBorder{
			Top:    ".-=~=-.",
			Mid:    []string{"( _ __)", "(__  _)"},
			Bottom: "`-._.-'",
		},
		MidBorder: MapBorder{
			Top:    "-._.-=",
			Bottom: "-._.-=",
		},
		RightBorder: MapBorder{
			Top:    ".-=~=-.",
			Mid:    []string{"( _ __)", "(__  _)"},
			Bottom: "`-._.-'",
		},
	}

	// Apply any overrides
	for sym, txt := range legendOverrides {
		if _, ok := legend[sym]; ok {
			legend[sym] = txt
		}
	}

	for i := 0; i < mapData.Height; i++ {
		for sym, txt := range legend {
			if strings.Contains(mapData.DisplayLines[i], sym) {
				txtLc := strings.ToLower(txt)
				mapData.DisplayLines[i] = strings.Replace(mapData.DisplayLines[i], sym, fmt.Sprintf(`<ansi fg="map-room"><ansi fg="map-%s" bg="mapbg-%s">%s</ansi></ansi>`, txtLc, txtLc, sym), -1)
			}
		}
	}

	return mapData, nil
}
