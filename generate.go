// Package zen is the root for all ebiten-zen files
package zen

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"
)

// DungeonTile represents the type of tile being used
type DungeonTile int8

// DungeonTiles
const (
	DungeonTileVoid DungeonTile = iota
	DungeonTileWall
	DungeonTilePreWall // placeholder for walls during generation
	DungeonTileFloor

	DungeonTileDoor
	DungeonTileRoomBegin
	DungeonTileRoomEnd
)

// Tiles aliases for creating neat maps manually
const (
	V = DungeonTileVoid
	W = DungeonTileWall
	P = DungeonTilePreWall
	F = DungeonTileFloor
)

// DoorDirection specifies the direction of a Door
type DoorDirection int8

// Directions
const (
	DoorDirectionHorizontal DoorDirection = iota
	DoorDirectionVertical
)

func (t DungeonTile) String() string {
	switch t {
	case DungeonTileVoid:
		return "◾"
	case DungeonTilePreWall:
		return "🔳"
	case DungeonTileWall:
		return "⬜"
	case DungeonTileFloor:
		return "⬛"
	case DungeonTileDoor:
		return "🚪"
	case DungeonTileRoomBegin:
		return "🟢"
	case DungeonTileRoomEnd:
		return "🔴"
	}

	return "🚧"
}

// Dungeon represents the map, Tiles are stored in [y][x] order, but GetTile can be used with (x,y) order to simplify some
// processes
type Dungeon struct {
	Width, Height int

	Tiles [][]DungeonTile // indexed [y][x]
	Rooms map[Rect]struct{}
	Doors map[Rect]DoorDirection

	ShowErrorMessages bool

	startTime           time.Time // for generation retry
	DurationBeforeRetry time.Duration
	genStartTime        time.Time // for error
	DurationBeforeError time.Duration

	Border                    int // don't place tiles in this area
	WallThickness             int // how many tiles thick the walls are
	MinDoorSize               int
	MaxDoorSize               int
	AllowRandomCorridorOffset bool // 
	MaxRoomWidth              int
	MaxRoomHeight             int
	MinRoomWidth              int
	MinRoomHeight             int
	MinIslandSize             int // RandomWalk only; any TileVoid islands < this are filled with TileFloor
}

var (
	rng *rand.Rand
	// ErrOutOfBounds is returned when a tile is attempted to be placed out of bounds
	ErrOutOfBounds = errors.New("Coordinate out of bounds")
	// ErrNotEnoughSpace is returned when there isn't enough space to generate the dungeon
	ErrNotEnoughSpace = errors.New("Not enough space to generate dungeon")
	// ErrGenerationTimeout is returned when generation has deadlocked
	ErrGenerationTimeout = errors.New("Took too long to generate dungeon")
	// ErrFloorAlreadyPlaced is returned when a floor tile is already placed
	ErrFloorAlreadyPlaced = errors.New("Floor tile already placed")
)

// ResetDungeon clears the tiles from the dungeon
func (dungeon *Dungeon) ResetDungeon(width, height int) {
	tiles := make([][]DungeonTile, height)
	for i := range tiles {
		tiles[i] = make([]DungeonTile, width)
	}
	dungeon.Tiles = tiles

	dungeon.Rooms = make(map[Rect]struct{})
	dungeon.Doors = make(map[Rect]DoorDirection)
}

// NewDungeon returns a new dungeon instance
func NewDungeon(width, height int) *Dungeon {
	s1 := rand.NewSource(time.Now().UnixNano())
	rng = rand.New(s1)

	dungeon := &Dungeon{
		Width:  width,
		Height: height,

		ShowErrorMessages: false,

		startTime:           time.Now(),
		DurationBeforeRetry: time.Millisecond * 250,
		DurationBeforeError: time.Second,

		Border:                    2,
		WallThickness:             2,
		MinDoorSize:               1,
		MaxDoorSize:               1,
		AllowRandomCorridorOffset: false,
		MaxRoomWidth:              8,
		MaxRoomHeight:             8,
		MinRoomWidth:              4,
		MinRoomHeight:             4,
		MinIslandSize:             26,
	}
	dungeon.ResetDungeon(width, height)
	return dungeon
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func absInt(a int) int {
	if a < 0 {
		return a * -1
	}
	return a
}
func randInt(a, b int) int {
	return rng.Int()%(b+1-a) + a
}

// GetTile returns a tile
func (dungeon *Dungeon) GetTile(x, y int) (DungeonTile, error) {
	w, h, b := dungeon.Width, dungeon.Height, dungeon.Border
	if x >= w-b || x < 0+b || y >= h-b || y < 0+b {
		return DungeonTileVoid, ErrOutOfBounds
	}
	return dungeon.Tiles[y][x], nil
}

// SetTile sets a tile
func (dungeon *Dungeon) SetTile(x, y int, t DungeonTile) error {
	w, h, b := dungeon.Width, dungeon.Height, dungeon.Border
	if t == DungeonTileFloor && (x >= w-b || x < 0+b || y >= h-b || y < 0+b) {
		return ErrOutOfBounds
	}

	dungeon.Tiles[y][x] = t
	return nil
}

// AddWalls adds a TileWall around every TileFloor
func (dungeon *Dungeon) AddWalls() {
	w, h, t := dungeon.Width, dungeon.Height, dungeon.WallThickness
	b := dungeon.Border
	dungeon.Border = 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if tile, err := dungeon.GetTile(x, y); err == nil {
				switch tile {
				case DungeonTileFloor:
					for dx := -t; dx <= t; dx++ {
						for dy := -t; dy <= t; dy++ {
							if tile, err := dungeon.GetTile(x+dx, y+dy); err == nil && tile == DungeonTileVoid {
								dungeon.SetTile(x+dx, y+dy, DungeonTileWall)
							}
						}
					}
				case DungeonTilePreWall:
					dungeon.SetTile(x, y, DungeonTileWall)
				}
			}
		}
	}
	dungeon.Border = b
}

func (dungeon *Dungeon) countSurrounding(x, y int, checkType DungeonTile) int {
	var count int
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if !(dx == 0 && dy == 0) {
				if tile, err := dungeon.GetTile(x+dx, y+dy); err == nil && tile == checkType {
					count++
				}
			}
		}
	}
	return count
}

func (dungeon *Dungeon) countSurroundingPolar(x, y int, checkType DungeonTile) int {
	var count int
	if tile, err := dungeon.GetTile(x+1, y); err == nil && tile == checkType {
		count++
	}
	if tile, err := dungeon.GetTile(x-1, y); err == nil && tile == checkType {
		count++
	}
	if tile, err := dungeon.GetTile(x, y+1); err == nil && tile == checkType {
		count++
	}
	if tile, err := dungeon.GetTile(x, y-1); err == nil && tile == checkType {
		count++
	}
	return count
}

func (dungeon *Dungeon) countIslandPolar(x, y int, checkType DungeonTile) (int, map[Rect]struct{}) {
	// recursively count neigboring tiles
	m := make(map[Rect]struct{})
	var g func(x, y int)
	g = func(x, y int) {
		if tile, err := dungeon.GetTile(x, y); err == nil && tile == checkType {
			c := Rect{X: x, Y: y}
			if _, ok := m[c]; !ok {
				m[c] = struct{}{}
				g(x+1, y)
				g(x-1, y)
				g(x, y+1)
				g(x, y-1)
			}
		}
	}
	g(x, y)
	var count int
	for range m {
		count++
	}
	return count, m
}

// CleanWalls replaces walls which don't have mustSurroundCount walls around them
func (dungeon *Dungeon) CleanWalls(mustSurroundCount int) {
	w, h := dungeon.Width, dungeon.Height
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if tile, err := dungeon.GetTile(x, y); err == nil && tile == DungeonTileWall {
				if dungeon.countSurrounding(x, y, DungeonTileFloor) >= mustSurroundCount {
					dungeon.SetTile(x, y, DungeonTileFloor)
				}
			}
		}
	}
}

// CleanIslands removes the pockets of WallVoids floating in the sea of WallFloors
func (dungeon *Dungeon) CleanIslands() {
	// Find islands
	islands := make([]map[Rect]struct{}, 0)
	for x := 0; x < dungeon.Width; x++ {
		for y := 0; y < dungeon.Height; y++ {
			var found bool
			for _, i := range islands {
				c := Rect{X: x, Y: y}
				if _, ok := i[c]; ok {
					found = true
				}
			}
			if !found {
				c, m := dungeon.countIslandPolar(x, y, DungeonTileVoid)
				islands = append(islands, m)
				// Remove island
				if c < dungeon.MinIslandSize {
					for co := range m {
						dungeon.SetTile(co.X, co.Y, DungeonTileFloor)
					}
				}
			}
		}
	}
}

// GenerateRandomWalk generates the dungeon using the random walk function
// The dungeon will look chaotic yet natural and all tiles will be touching each other
// dungeon.Convexity, dungeon.WallThickness and dungeon.CorridorSize is used
// Ensure that tileCount isn't too high or else dungeon generation can take a while
func (dungeon *Dungeon) GenerateRandomWalk(tileCount int) error {
	dungeon.genStartTime = time.Now()

	w, h := dungeon.Width, dungeon.Height

	var g func() error
	g = func() error {
		dungeon.ResetDungeon(dungeon.Width, dungeon.Height)
		x, y := w/2, h/2
		minX, maxX, minY, maxY := w, 0, h, 0
		var dx, dy int

		for tc := 0; tc < tileCount; {
			if time.Now().Sub(dungeon.genStartTime) > dungeon.DurationBeforeError {
				return ErrGenerationTimeout
			} else if time.Now().Sub(dungeon.startTime) > dungeon.DurationBeforeRetry {
				if dungeon.ShowErrorMessages {
					log.Println("Timeout, retrying gen")
				}
				return g()
			}

			switch rng.Int() % 8 {
			case 0:
				dx = -1
				dy = 0
			case 1:
				dx = 1
				dy = 0
			case 2:
				dx = 0
				dy = -1
			case 3:
				dx = 0
				dy = 1
			default:
				// use the same direction as last time
			}
			x += dx
			y += dy

			cs := randInt(dungeon.MinDoorSize, dungeon.MaxDoorSize)
			for tx := x - cs/2; tx < x+cs/2; tx++ {
				for ty := y - cs/2; ty < y+cs/2; ty++ {
					tc++
					if tile, err := dungeon.GetTile(tx, ty); err == nil && tile != DungeonTileVoid {
						tc--
					} else if dungeon.SetTile(tx, ty, DungeonTileFloor) == ErrOutOfBounds {
						x = w / 2
						y = h / 2
						tc--
						goto cont
					}
				}
			}

			minX = minInt(minX, x)
			maxX = maxInt(maxX, x)
			minY = minInt(minY, y)
			maxY = maxInt(maxY, y)
		cont:
		}

		// Check convexity
		var convX bool
		cy := minY + (maxY-minY)/2
		var foundFloor, inGap bool
		for cx := minX; cx < maxX; cx++ {
			if tile, err := dungeon.GetTile(cx, cy); err == nil {
				switch tile {
				case DungeonTileFloor:
					if foundFloor && inGap {
						convX = true
						goto done
					}
					foundFloor = true
				case DungeonTileVoid:
					if foundFloor {
						inGap = true
					}
				}
			}
		}
	done:
		if !convX {
			if dungeon.ShowErrorMessages {
				log.Println("no convexity, retrying gen")
			}
			return g()
		}

		return nil
	}

	return g()
}

// Rect is used for storing the x,y,w,h of a room or corridor
type Rect struct {
	X, Y int
	W, H int
}

// GenerateDungeonGrid generates the dungeon using the dungeon grid function
// The dungeon will look neat, with rooms aligned perfectly in a grid. dungeon.MaxRoomWidth is used for both the width and
// the height of the rooms as all rooms are the same size and shape.
// dungeon.WallThickness, dungeon.MaxRoomWidth and dungeon.CorridorSize and dungeon.AllowRandomCorridorOffset are used
func (dungeon *Dungeon) GenerateDungeonGrid(roomCount int) error {
	dungeon.genStartTime = time.Now()

	s := dungeon.MaxRoomWidth
	mw := (dungeon.Width-dungeon.Border*2)/(s+dungeon.WallThickness) + 1
	mh := (dungeon.Height-dungeon.Border*2)/(s+dungeon.WallThickness) + 1

	if dungeon.ShowErrorMessages {
		fmt.Printf("Max grid size is %d x %d, so max roomCount is %d. Use fewer rooms for a better result.\n", mw-1, mh-1, (mw-1)*(mh-1))
	}

	// if roomCount > (mw-2)*(mh-2) {
	// 	return ErrNotEnoughSpace
	// }

	var g func() error
	g = func() error {
		dungeon.ResetDungeon(dungeon.Width, dungeon.Height)
		sx, sy := int(mw/2), int(mh/2)
		dungeon.startTime = time.Now()
		// Create rooms layout data structure
		rooms := make([][]bool, mh)
		for i := range rooms {
			rooms[i] = make([]bool, mw)
		}

		previousRooms := make([][]Rect, 1)
		for rc := roomCount; rc > 0; rc-- {
			if time.Now().Sub(dungeon.genStartTime) > dungeon.DurationBeforeError {
				return ErrGenerationTimeout
			} else if time.Now().Sub(dungeon.startTime) > dungeon.DurationBeforeRetry {
				if dungeon.ShowErrorMessages {
					log.Println("Timeout, retrying gen")
				}
				return g()
			}
			switch rng.Int() % 4 {
			case 0:
				sx--
			case 1:
				sx++
			case 2:
				sy--
			case 3:
				sy++
			}

			countAdj := func(iy, ix int) int {
				var count int
				if iy > 0 && rooms[iy-1][ix] {
					count++
				}
				if iy+1 < mh && rooms[iy+1][ix] {
					count++
				}
				if ix > 0 && rooms[iy][ix-1] {
					count++
				}
				if ix+1 < mw && rooms[iy][ix+1] {
					count++
				}
				return count
			}

			if sx >= mw || sx <= 0 || sy >= mh || sy <= 0 || (countAdj(sy, sx) >= 2 && rooms[sy][sx]) {
				rc++
				for l := 0; l < len(previousRooms); l++ {
					for i := 0; i < len(previousRooms[l]); i++ { // start from beginning
						roomCoord := previousRooms[l][i]
						rc := countAdj(roomCoord.Y, roomCoord.X)
						if rc >= 0 && rc <= 2 {
							sx = roomCoord.X
							sy = roomCoord.Y
							previousRooms = append(previousRooms, make([]Rect, 0))
							goto good
						}
					}
				}
				return ErrNotEnoughSpace
			}
		good:
			// Append room coord for rewinding purposes
			rooms[sy][sx] = true
			previousRooms[len(previousRooms)-1] = append(previousRooms[len(previousRooms)-1], Rect{X: sx, Y: sy})
		}

		for pr := 0; pr < len(previousRooms); pr++ {
			// log.Println(previousRooms[pr])
			for i, cur := range previousRooms[pr] {
				sy, sx = cur.Y, cur.X
				room := Rect{
					X: sx*s + sx*dungeon.WallThickness - dungeon.MaxRoomWidth,
					Y: sy*s + sy*dungeon.WallThickness - dungeon.MaxRoomWidth,
					W: dungeon.MaxRoomWidth,
					H: dungeon.MaxRoomWidth,
				}
				dungeon.Rooms[room] = struct{}{}

				// Fill in the dungeon's tiles with the room
				for dx := room.X; dx < room.X+room.W; dx++ {
					for dy := room.Y; dy < room.Y+room.H; dy++ {
						dungeon.SetTile(dx, dy, DungeonTileFloor)
					}
				}

				if i == 0 {
					continue
				}

				// Corridors
				prev := previousRooms[pr][i-1]
				dx, dy := cur.X-prev.X, cur.Y-prev.Y
				x1 := prev.X*s - dungeon.MaxRoomWidth/2
				x2 := cur.X*s - dungeon.MaxRoomWidth/2
				y1 := prev.Y*s - dungeon.MaxRoomWidth/2
				y2 := cur.Y*s - dungeon.MaxRoomWidth/2
				cd := DoorDirectionHorizontal
				cs := randInt(dungeon.MinDoorSize, dungeon.MaxDoorSize)
				var offsetCy, offsetCx int
				if dungeon.AllowRandomCorridorOffset {
					offsetCy = (dungeon.MaxRoomWidth - cs)
					offsetCy = randInt(-offsetCy/2, offsetCy/2)
					offsetCx = (dungeon.MaxRoomWidth - cs)
					offsetCx = randInt(-offsetCx/2, offsetCx/2)
				}
				switch {
				case dx == -1: // left
					x1 = x2 + dungeon.MaxRoomWidth/2 + dungeon.MaxRoomWidth%2
					x2 = x1 + dungeon.WallThickness
					y1 = y1 - cs/2 - offsetCy
					y2 = y2 + cs/2 + cs%2 - offsetCy
					cd = DoorDirectionVertical
				case dx == 1:
					x1 = x2 - dungeon.MaxRoomWidth/2 - dungeon.WallThickness
					x2 = x1 + dungeon.WallThickness
					y1 = y1 - cs/2 - offsetCy
					y2 = y2 + cs/2 + cs%2 - offsetCy
					cd = DoorDirectionVertical
				case dy == -1:
					y1 = y2 + dungeon.MaxRoomWidth/2 + dungeon.MaxRoomWidth%2
					y2 = y1 + dungeon.WallThickness
					x1 = x1 - cs/2 - offsetCx
					x2 = x2 + cs/2 + cs%2 - offsetCx
				case dy == 1:
					y1 = y2 - dungeon.MaxRoomWidth/2 - dungeon.WallThickness
					y2 = y1 + dungeon.WallThickness
					x1 = x1 - cs/2 - offsetCx
					x2 = x2 + cs/2 + cs%2 - offsetCx
				default:
					if dungeon.ShowErrorMessages {
						log.Println("somehow, dx,dy > abs 1", cur, prev, dx, dy)
					}
				}

				cx := Rect{
					X: x1 + sx*dungeon.WallThickness,
					Y: y1 + sy*dungeon.WallThickness,
					W: x2 - x1,
					H: y2 - y1,
				}
				if dungeon.WallThickness > 1 {
					switch cd {
					case DoorDirectionHorizontal:
						cx.H = 1
						cx.Y += (dungeon.WallThickness/2 + dungeon.WallThickness%2) - 1
					case DoorDirectionVertical:
						cx.W = 1
						cx.X += (dungeon.WallThickness/2 + dungeon.WallThickness%2) - 1
					}
				}
				dungeon.Doors[cx] = cd
				for x := x1; x < x2; x++ {
					for y := y1; y < y2; y++ {
						dungeon.SetTile(x+sx*dungeon.WallThickness, y+sy*dungeon.WallThickness, DungeonTileFloor)
					}
				}
			}
		}
		return nil
	}
	return g()
}

// GenerateDungeon generates the dungeon using a more fluid algorithm
// The dungeon will have randomly sized rooms
// dungeon.WallThickness, dungeon.MinRoomWidth|Height, dungeon.MaxRoomWidth|Height, dungeon.CorridorSize and
// dungeon.AllowRandomCorridorOffset are used
func (dungeon *Dungeon) GenerateDungeon(roomCount int) error {
	dungeon.genStartTime = time.Now()

	s := dungeon.MaxRoomWidth
	mw := (dungeon.Width - dungeon.Border*2) / s
	mh := (dungeon.Height - dungeon.Border*2) / s

	if roomCount > (mw-2)*(mh-2) {
		return ErrNotEnoughSpace
	}

	var g func() error
	g = func() error {
		dungeon.ResetDungeon(dungeon.Width, dungeon.Height)
		dungeon.startTime = time.Now()
		// Helper func to place rooms
		placeRoom := func(x, y, w, h int) error {
			// Check area
			for dx := x - dungeon.WallThickness; dx < x+w+dungeon.WallThickness; dx++ {
				for dy := y - dungeon.WallThickness; dy < y+h+dungeon.WallThickness; dy++ {
					if tile, err := dungeon.GetTile(dx, dy); err == nil && tile == DungeonTileFloor {
						return ErrFloorAlreadyPlaced
					} else if err != nil {
						return err
					}
				}
			}
			// Place
			for dx := x - dungeon.WallThickness; dx < x+w+dungeon.WallThickness; dx++ {
				for dy := y - dungeon.WallThickness; dy < y+h+dungeon.WallThickness; dy++ {
					if dx < x || dx > x+w-1 || dy < y || dy > y+h-1 {
						// Temp wall
						if tile, err := dungeon.GetTile(dx, dy); err == nil && tile == DungeonTileVoid {
							if err := dungeon.SetTile(dx, dy, DungeonTilePreWall); err != nil {
								return err
							}
						}
					} else {
						// Floor
						if err := dungeon.SetTile(dx, dy, DungeonTileFloor); err != nil {
							return err
						}
					}
				}
			}
			// Set dungeon.Rooms
			dungeon.Rooms[Rect{
				X: x,
				Y: y,
				W: w,
				H: h,
			}] = struct{}{}
			return nil
		}

		// Random first room size
		sx, sy := dungeon.Width/2, dungeon.Height/2
		rw := randInt(dungeon.MinRoomWidth, dungeon.MaxRoomWidth)
		rh := randInt(dungeon.MinRoomHeight, dungeon.MaxRoomHeight)

		// Place the first room into the dungeon
		placeRoom(sx, sy, rw, rh)

		previousRooms := make([]Rect, 0)
		previousRooms = append(previousRooms, Rect{X: sx, Y: sy, W: rw, H: rh})

		for rc := roomCount - 1; rc > 0; rc-- {
			if time.Now().Sub(dungeon.genStartTime) > dungeon.DurationBeforeError {
				return ErrGenerationTimeout
			} else if time.Now().Sub(dungeon.startTime) > dungeon.DurationBeforeRetry {
				if dungeon.ShowErrorMessages {
					log.Println("Timeout, retrying gen")
				}
				return g()
			}

			// Offset position by last room
			osx := sx
			osy := sy
			orw := rw
			orh := rh
			rw = randInt(dungeon.MinRoomWidth, dungeon.MaxRoomWidth)
			rh = randInt(dungeon.MinRoomHeight, dungeon.MaxRoomHeight)
			cx, cy := osx, osy // corridor position
			cs := randInt(dungeon.MinDoorSize, dungeon.MaxDoorSize)
			var cw, ch int
			var offsetCy, offsetCx int
			if dungeon.AllowRandomCorridorOffset {
				offsetCy = (minInt(rh, orh) - ch)
				offsetCy = randInt(-cs/2, offsetCy/2-cs/2)
				offsetCx = (minInt(rw, orw) - cw)
				offsetCx = randInt(-cs/2, offsetCx/2-cs/2)
			}
			cd := DoorDirectionHorizontal
			switch rng.Int() % 4 {
			case 0: // left
				cw = dungeon.WallThickness
				ch = cs
				sx = sx - dungeon.WallThickness - rw
				cx = sx + rw
				cy = cy + (ch / 2) + offsetCy
				cd = DoorDirectionVertical
			case 1: // right
				cw = dungeon.WallThickness
				ch = cs
				sx = sx + orw + dungeon.WallThickness
				cx = sx - dungeon.WallThickness
				cy = cy + (ch / 2) + offsetCy
				cd = DoorDirectionVertical
			case 2: // up
				cw = cs
				ch = dungeon.WallThickness
				sy = sy - dungeon.WallThickness - rh
				cy = sy + rh
				cx = cx + (cw / 2) + offsetCx
			case 3: // down
				cw = cs
				ch = dungeon.WallThickness
				sy = sy + orh + dungeon.WallThickness
				cy = sy - dungeon.WallThickness
				cx = cx + (cw / 2) + offsetCx
			}

			if err := placeRoom(sx, sy, rw, rh); err != nil {
				if dungeon.ShowErrorMessages {
					log.Println("rollback:", err, sx, sy, rw, rh)
				}
				c := previousRooms[rng.Int()%len(previousRooms)]
				sx = c.X
				sy = c.Y
				rw = c.W
				rh = c.H
				rc++
				continue
			}

			// Corridors
			door := Rect{
				X: cx,
				Y: cy,
				W: cw,
				H: ch,
			}
			if dungeon.WallThickness > 1 {
				switch cd {
				case DoorDirectionHorizontal:
					door.H = 1
					door.Y += (dungeon.WallThickness/2 + dungeon.WallThickness%2) - 1
				case DoorDirectionVertical:
					door.W = 1
					door.X += (dungeon.WallThickness/2 + dungeon.WallThickness%2) - 1
				}
			}
			dungeon.Doors[door] = cd
			for x := cx; x < cx+cw; x++ {
				for y := cy; y < cy+ch; y++ {
					dungeon.SetTile(x, y, DungeonTileFloor)
				}
			}

			previousRooms = append(previousRooms, Rect{X: sx, Y: sy, W: rw, H: rh})
		}

		return nil
	}
	return g()
}
