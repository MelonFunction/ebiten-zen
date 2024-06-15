// Package zen is the root for all ebiten-zen files
package zen

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"
)

// Tile represents the type of tile being used
type Tile int8

// Tiles
const (
	TileVoid Tile = iota
	TileWall
	TilePreWall // placeholder for walls during generation
	TileFloor

	TileDoor
	TileRoomBegin
	TileRoomEnd
)

// Tiles aliases for creating neat maps manually
const (
	V = TileVoid
	W = TileWall
	P = TilePreWall
	F = TileFloor
)

// DoorDirection specifies the direction of a Door
type DoorDirection int8

// Directions
const (
	DoorDirectionHorizontal DoorDirection = iota
	DoorDirectionVertical
)

func (t Tile) String() string {
	switch t {
	case TileVoid:
		return "◾"
	case TilePreWall:
		return "🔳"
	case TileWall:
		return "⬜"
	case TileFloor:
		return "⬛"
	case TileDoor:
		return "🚪"
	case TileRoomBegin:
		return "🟢"
	case TileRoomEnd:
		return "🔴"
	}

	return "🚧"
}

// World represents the map, Tiles are stored in [y][x] order, but GetTile can be used with (x,y) order to simplify some
// processes
type World struct {
	Width, Height int

	Tiles [][]Tile // indexed [y][x]
	Rooms map[Rect]struct{}
	Doors map[Rect]DoorDirection

	ShowErrorMessages bool

	startTime           time.Time // for generation retry
	DurationBeforeRetry time.Duration
	genStartTime        time.Time // for error
	DurationBeforeError time.Duration

	Border                    int // don't place tiles in this area
	WallThickness             int // how many tiles thick the walls are
	MinDoorSize               int // how wide a door/corridor is
	MaxDoorSize               int
	AllowRandomCorridorOffset bool // places door at random position in the wall
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
	// ErrNotEnoughSpace is returned when there isn't enough space to generate the world
	ErrNotEnoughSpace = errors.New("Not enough space to generate world")
	// ErrGenerationTimeout is returned when generation has deadlocked
	ErrGenerationTimeout = errors.New("Took too long to generate world")
	// ErrFloorAlreadyPlaced is returned when a floor tile is already placed
	ErrFloorAlreadyPlaced = errors.New("Floor tile already placed")
)

// Reset clears the tiles from the world
func (world *World) Reset(width, height int) {
	tiles := make([][]Tile, height)
	for i := range tiles {
		tiles[i] = make([]Tile, width)
	}
	world.Tiles = tiles

	world.Rooms = make(map[Rect]struct{})
	world.Doors = make(map[Rect]DoorDirection)
}

// NewWorld returns a new world instance
func NewWorld(width, height int) *World {
	s1 := rand.NewSource(time.Now().UnixNano())
	rng = rand.New(s1)

	world := &World{
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
	world.Reset(width, height)
	return world
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
func (world *World) GetTile(x, y int) (Tile, error) {
	w, h, b := world.Width, world.Height, world.Border
	if x >= w-b || x < 0+b || y >= h-b || y < 0+b {
		return TileVoid, ErrOutOfBounds
	}
	return world.Tiles[y][x], nil
}

// SetTile sets a tile
func (world *World) SetTile(x, y int, t Tile) error {
	w, h, b := world.Width, world.Height, world.Border
	if t == TileFloor && (x >= w-b || x < 0+b || y >= h-b || y < 0+b) {
		return ErrOutOfBounds
	}

	world.Tiles[y][x] = t
	return nil
}

// AddWalls adds a TileWall around every TileFloor
func (world *World) AddWalls() {
	w, h, t := world.Width, world.Height, world.WallThickness
	b := world.Border
	world.Border = 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if tile, err := world.GetTile(x, y); err == nil {
				switch tile {
				case TileFloor:
					for dx := -t; dx <= t; dx++ {
						for dy := -t; dy <= t; dy++ {
							if tile, err := world.GetTile(x+dx, y+dy); err == nil && tile == TileVoid {
								world.SetTile(x+dx, y+dy, TileWall)
							}
						}
					}
				case TilePreWall:
					world.SetTile(x, y, TileWall)
				}
			}
		}
	}
	world.Border = b
}

func (world *World) countSurrounding(x, y int, checkType Tile) int {
	var count int
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if !(dx == 0 && dy == 0) {
				if tile, err := world.GetTile(x+dx, y+dy); err == nil && tile == checkType {
					count++
				}
			}
		}
	}
	return count
}

func (world *World) countSurroundingPolar(x, y int, checkType Tile) int {
	var count int
	if tile, err := world.GetTile(x+1, y); err == nil && tile == checkType {
		count++
	}
	if tile, err := world.GetTile(x-1, y); err == nil && tile == checkType {
		count++
	}
	if tile, err := world.GetTile(x, y+1); err == nil && tile == checkType {
		count++
	}
	if tile, err := world.GetTile(x, y-1); err == nil && tile == checkType {
		count++
	}
	return count
}

func (world *World) countIslandPolar(x, y int, checkType Tile) (int, map[Rect]struct{}) {
	// recursively count neigboring tiles
	m := make(map[Rect]struct{})
	var g func(x, y int)
	g = func(x, y int) {
		if tile, err := world.GetTile(x, y); err == nil && tile == checkType {
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
func (world *World) CleanWalls(mustSurroundCount int) {
	w, h := world.Width, world.Height
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if tile, err := world.GetTile(x, y); err == nil && tile == TileWall {
				if world.countSurrounding(x, y, TileFloor) >= mustSurroundCount {
					world.SetTile(x, y, TileFloor)
				}
			}
		}
	}
}

// CleanIslands removes the pockets of WallVoids floating in the sea of WallFloors
func (world *World) CleanIslands() {
	// Find islands
	islands := make([]map[Rect]struct{}, 0)
	for x := 0; x < world.Width; x++ {
		for y := 0; y < world.Height; y++ {
			var found bool
			for _, i := range islands {
				c := Rect{X: x, Y: y}
				if _, ok := i[c]; ok {
					found = true
				}
			}
			if !found {
				c, m := world.countIslandPolar(x, y, TileVoid)
				islands = append(islands, m)
				// Remove island
				if c < world.MinIslandSize {
					for co := range m {
						world.SetTile(co.X, co.Y, TileFloor)
					}
				}
			}
		}
	}
}

// GenerateRandomWalk generates the world using the random walk function
// The world will look chaotic yet natural and all tiles will be touching each other
// world.Convexity, world.WallThickness and world.CorridorSize is used
// Ensure that tileCount isn't too high or else world generation can take a while
func (world *World) GenerateRandomWalk(tileCount int) error {
	world.genStartTime = time.Now()

	w, h := world.Width, world.Height

	var g func() error
	g = func() error {
		world.Reset(world.Width, world.Height)
		x, y := w/2, h/2
		minX, maxX, minY, maxY := w, 0, h, 0
		var dx, dy int

		for tc := 0; tc < tileCount; {
			if time.Now().Sub(world.genStartTime) > world.DurationBeforeError {
				return ErrGenerationTimeout
			} else if time.Now().Sub(world.startTime) > world.DurationBeforeRetry {
				if world.ShowErrorMessages {
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

			cs := randInt(world.MinDoorSize, world.MaxDoorSize)
			for tx := x - cs/2; tx < x+cs/2; tx++ {
				for ty := y - cs/2; ty < y+cs/2; ty++ {
					tc++
					if tile, err := world.GetTile(tx, ty); err == nil && tile != TileVoid {
						tc--
					} else if world.SetTile(tx, ty, TileFloor) == ErrOutOfBounds {
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
			if tile, err := world.GetTile(cx, cy); err == nil {
				switch tile {
				case TileFloor:
					if foundFloor && inGap {
						convX = true
						goto done
					}
					foundFloor = true
				case TileVoid:
					if foundFloor {
						inGap = true
					}
				}
			}
		}
	done:
		if !convX {
			if world.ShowErrorMessages {
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

// GenerateDungeonGrid generates the world using the world grid function
// The world will look neat, with rooms aligned perfectly in a grid. world.MaxRoomWidth is used for both the width and
// the height of the rooms as all rooms are the same size and shape.
// world.WallThickness, world.MaxRoomWidth and world.CorridorSize and world.AllowRandomCorridorOffset are used
func (world *World) GenerateDungeonGrid(roomCount int) error {
	world.genStartTime = time.Now()

	s := world.MaxRoomWidth
	mw := (world.Width-world.Border*2)/(s+world.WallThickness) + 1
	mh := (world.Height-world.Border*2)/(s+world.WallThickness) + 1

	if world.ShowErrorMessages {
		fmt.Printf("Max grid size is %d x %d, so max roomCount is %d. Use fewer rooms for a better result.\n", mw-1, mh-1, (mw-1)*(mh-1))
	}

	// if roomCount > (mw-2)*(mh-2) {
	// 	return ErrNotEnoughSpace
	// }

	var g func() error
	g = func() error {
		world.Reset(world.Width, world.Height)
		sx, sy := int(mw/2), int(mh/2)
		world.startTime = time.Now()
		// Create rooms layout data structure
		rooms := make([][]bool, mh)
		for i := range rooms {
			rooms[i] = make([]bool, mw)
		}

		previousRooms := make([][]Rect, 1)
		for rc := roomCount; rc > 0; rc-- {
			if time.Now().Sub(world.genStartTime) > world.DurationBeforeError {
				return ErrGenerationTimeout
			} else if time.Now().Sub(world.startTime) > world.DurationBeforeRetry {
				if world.ShowErrorMessages {
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
					X: sx*s + sx*world.WallThickness - world.MaxRoomWidth,
					Y: sy*s + sy*world.WallThickness - world.MaxRoomWidth,
					W: world.MaxRoomWidth,
					H: world.MaxRoomWidth,
				}
				world.Rooms[room] = struct{}{}

				// Fill in the world's tiles with the room
				for dx := room.X; dx < room.X+room.W; dx++ {
					for dy := room.Y; dy < room.Y+room.H; dy++ {
						world.SetTile(dx, dy, TileFloor)
					}
				}

				if i == 0 {
					continue
				}

				// Corridors
				prev := previousRooms[pr][i-1]
				dx, dy := cur.X-prev.X, cur.Y-prev.Y
				x1 := prev.X*s - world.MaxRoomWidth/2
				x2 := cur.X*s - world.MaxRoomWidth/2
				y1 := prev.Y*s - world.MaxRoomWidth/2
				y2 := cur.Y*s - world.MaxRoomWidth/2
				cd := DoorDirectionHorizontal
				cs := randInt(world.MinDoorSize, world.MaxDoorSize)
				var offsetCy, offsetCx int
				if world.AllowRandomCorridorOffset {
					offsetCy = (world.MaxRoomWidth - cs)
					offsetCy = randInt(-offsetCy/2, offsetCy/2)
					offsetCx = (world.MaxRoomWidth - cs)
					offsetCx = randInt(-offsetCx/2, offsetCx/2)
				}
				switch {
				case dx == -1: // left
					x1 = x2 + world.MaxRoomWidth/2 + world.MaxRoomWidth%2
					x2 = x1 + world.WallThickness
					y1 = y1 - cs/2 - offsetCy
					y2 = y2 + cs/2 + cs%2 - offsetCy
					cd = DoorDirectionVertical
				case dx == 1:
					x1 = x2 - world.MaxRoomWidth/2 - world.WallThickness
					x2 = x1 + world.WallThickness
					y1 = y1 - cs/2 - offsetCy
					y2 = y2 + cs/2 + cs%2 - offsetCy
					cd = DoorDirectionVertical
				case dy == -1:
					y1 = y2 + world.MaxRoomWidth/2 + world.MaxRoomWidth%2
					y2 = y1 + world.WallThickness
					x1 = x1 - cs/2 - offsetCx
					x2 = x2 + cs/2 + cs%2 - offsetCx
				case dy == 1:
					y1 = y2 - world.MaxRoomWidth/2 - world.WallThickness
					y2 = y1 + world.WallThickness
					x1 = x1 - cs/2 - offsetCx
					x2 = x2 + cs/2 + cs%2 - offsetCx
				default:
					if world.ShowErrorMessages {
						log.Println("somehow, dx,dy > abs 1", cur, prev, dx, dy)
					}
				}

				cx := Rect{
					X: x1 + sx*world.WallThickness,
					Y: y1 + sy*world.WallThickness,
					W: x2 - x1,
					H: y2 - y1,
				}
				if world.WallThickness > 1 {
					switch cd {
					case DoorDirectionHorizontal:
						cx.H = 1
						cx.Y += (world.WallThickness/2 + world.WallThickness%2) - 1
					case DoorDirectionVertical:
						cx.W = 1
						cx.X += (world.WallThickness/2 + world.WallThickness%2) - 1
					}
				}
				world.Doors[cx] = cd
				for x := x1; x < x2; x++ {
					for y := y1; y < y2; y++ {
						world.SetTile(x+sx*world.WallThickness, y+sy*world.WallThickness, TileFloor)
					}
				}
			}
		}
		return nil
	}
	return g()
}

// GenerateDungeon generates the world using a more fluid algorithm
// The world will have randomly sized rooms
// world.WallThickness, world.MinRoomWidth|Height, world.MaxRoomWidth|Height, world.CorridorSize and
// world.AllowRandomCorridorOffset are used
func (world *World) GenerateDungeon(roomCount int) error {
	world.genStartTime = time.Now()

	s := world.MaxRoomWidth
	mw := (world.Width - world.Border*2) / s
	mh := (world.Height - world.Border*2) / s

	if roomCount > (mw-2)*(mh-2) {
		return ErrNotEnoughSpace
	}

	var g func() error
	g = func() error {
		world.Reset(world.Width, world.Height)
		world.startTime = time.Now()
		// Helper func to place rooms
		placeRoom := func(x, y, w, h int) error {
			// Check area
			for dx := x - world.WallThickness; dx < x+w+world.WallThickness; dx++ {
				for dy := y - world.WallThickness; dy < y+h+world.WallThickness; dy++ {
					if tile, err := world.GetTile(dx, dy); err == nil && tile == TileFloor {
						return ErrFloorAlreadyPlaced
					} else if err != nil {
						return err
					}
				}
			}
			// Place
			for dx := x - world.WallThickness; dx < x+w+world.WallThickness; dx++ {
				for dy := y - world.WallThickness; dy < y+h+world.WallThickness; dy++ {
					if dx < x || dx > x+w-1 || dy < y || dy > y+h-1 {
						// Temp wall
						if tile, err := world.GetTile(dx, dy); err == nil && tile == TileVoid {
							if err := world.SetTile(dx, dy, TilePreWall); err != nil {
								return err
							}
						}
					} else {
						// Floor
						if err := world.SetTile(dx, dy, TileFloor); err != nil {
							return err
						}
					}
				}
			}
			// Set world.Rooms
			world.Rooms[Rect{
				X: x,
				Y: y,
				W: w,
				H: h,
			}] = struct{}{}
			return nil
		}

		// Random first room size
		sx, sy := world.Width/2, world.Height/2
		rw := randInt(world.MinRoomWidth, world.MaxRoomWidth)
		rh := randInt(world.MinRoomHeight, world.MaxRoomHeight)

		// Place the first room into the world
		placeRoom(sx, sy, rw, rh)

		previousRooms := make([]Rect, 0)
		previousRooms = append(previousRooms, Rect{X: sx, Y: sy, W: rw, H: rh})

		for rc := roomCount - 1; rc > 0; rc-- {
			if time.Now().Sub(world.genStartTime) > world.DurationBeforeError {
				return ErrGenerationTimeout
			} else if time.Now().Sub(world.startTime) > world.DurationBeforeRetry {
				if world.ShowErrorMessages {
					log.Println("Timeout, retrying gen")
				}
				return g()
			}

			// Offset position by last room
			osx := sx
			osy := sy
			orw := rw
			orh := rh
			rw = randInt(world.MinRoomWidth, world.MaxRoomWidth)
			rh = randInt(world.MinRoomHeight, world.MaxRoomHeight)
			cx, cy := osx, osy // corridor position
			cs := randInt(world.MinDoorSize, world.MaxDoorSize)
			var cw, ch int
			var offsetCy, offsetCx int
			if world.AllowRandomCorridorOffset {
				offsetCy = (minInt(rh, orh) - ch)
				offsetCy = randInt(-cs/2, offsetCy/2-cs/2)
				offsetCx = (minInt(rw, orw) - cw)
				offsetCx = randInt(-cs/2, offsetCx/2-cs/2)
			}
			cd := DoorDirectionHorizontal
			switch rng.Int() % 4 {
			case 0: // left
				cw = world.WallThickness
				ch = cs
				sx = sx - world.WallThickness - rw
				cx = sx + rw
				cy = cy + (ch / 2) + offsetCy
				cd = DoorDirectionVertical
			case 1: // right
				cw = world.WallThickness
				ch = cs
				sx = sx + orw + world.WallThickness
				cx = sx - world.WallThickness
				cy = cy + (ch / 2) + offsetCy
				cd = DoorDirectionVertical
			case 2: // up
				cw = cs
				ch = world.WallThickness
				sy = sy - world.WallThickness - rh
				cy = sy + rh
				cx = cx + (cw / 2) + offsetCx
			case 3: // down
				cw = cs
				ch = world.WallThickness
				sy = sy + orh + world.WallThickness
				cy = sy - world.WallThickness
				cx = cx + (cw / 2) + offsetCx
			}

			if err := placeRoom(sx, sy, rw, rh); err != nil {
				if world.ShowErrorMessages {
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
			if world.WallThickness > 1 {
				switch cd {
				case DoorDirectionHorizontal:
					door.H = 1
					door.Y += (world.WallThickness/2 + world.WallThickness%2) - 1
				case DoorDirectionVertical:
					door.W = 1
					door.X += (world.WallThickness/2 + world.WallThickness%2) - 1
				}
			}
			world.Doors[door] = cd
			for x := cx; x < cx+cw; x++ {
				for y := cy; y < cy+ch; y++ {
					world.SetTile(x, y, TileFloor)
				}
			}

			previousRooms = append(previousRooms, Rect{X: sx, Y: sy, W: rw, H: rh})
		}

		return nil
	}
	return g()
}
